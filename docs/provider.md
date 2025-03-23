# Gone V2 Provider 机制介绍

- [Gone V2 Provider 机制介绍](#gone-v2-provider-机制介绍)
  - [1. Gone 的依赖注入流程](#1-gone-的依赖注入流程)
  - [2. 不同的 Provider](#2-不同的-provider)
    - [2.1 按类型注入](#21-按类型注入)
    - [2.2 按名字注入](#22-按名字注入)
    - [2.3 基于名字的多类型 Provider](#23-基于名字的多类型-provider)
  - [3. “星号” Provider：`*`](#3-星号-provider)
  - [总结](#总结)


Gone V2 版本完全基于 Provider 机制实现依赖注入，其核心思想是通过 Provider 为对象提供依赖，而不必将所有第三方对象都包装为 `Goner`。下面将从 Gone 的依赖注入流程、Provider 的分类及其使用示例等方面进行详细说明。

---

## 1. Gone 的依赖注入流程

Gone 框架的依赖注入过程主要分为以下三个步骤：

1. **标记需要注入的对象**  
   对象需嵌入 `gone.Flag`，并对需要注入的字段使用 `gone` 标签。例如：
   ```go
   type UseConf struct {
   	gone.Flag
   	Name string `gone:"config,name"`
   	Int  string `gone:"config",int`
   	Dep  *Dep   `gone:"*,extend"`
   }
   ```
	- `gone` 标签格式：`gone:"${name},${extend}"`
		- `${name}`：指定 Provider 的名字或目标 `Goner` 的名字（当为 `*` 或省略时，表示按类型注入）。
		- `${extend}`：扩展参数，将传递给 Provider 的 `Provide` 方法（可选）。

2. **注册对象**  
   通过 `Load(goner Goner, options ...Option)` 将需要注入的对象注册到 Gone 框架中。

3. **自动注入和生命周期管理**  
   在框架启动过程中，自动对所有注册对象进行依赖注入，同时调用如 `BeforeInit`、`Init`、`Start`、`Stop` 等生命周期方法（如果对象实现了对应接口）。

> **注意：** Gone 框架只允许注册实现了 `Goner` 接口的对象，而通常通过嵌入 `gone.Flag` 来实现 `Goner` 接口。这样可以区分哪些对象需要注入、哪些不需要。但这也带来了一个问题：如何将第三方对象注入到框架中？为此，Provider 机制被引入。

---

## 2. 不同的 Provider

Gone 中主要支持两种注入方式：**按类型注入**和**按名字注入**。

### 2.1 按类型注入

当字段未指定特定名称时（标签中使用 `*` 或省略名称），框架会根据字段的类型查找合适的 Provider。这里定义了两种 Provider 接口：

- **支持传入参数的 Provider**
  ```go
  type Provider[T any] interface {
      Goner
      Provide(tagConf string) (T, error)
  }
  ```

- **无参数的 Provider**
  ```go
  type NoneParamProvider[T any] interface {
      Goner
      Provide() (T, error)
  }
  ```

**示例代码：**
```go
package main

import "github.com/gone-io/gone/v2"

type ThirdBusiness1 struct {
    Name string
}

type ThirdBusiness2 struct {
    Name string
}

// Provider 实现，带参数
type ThirdBusiness1Provider struct {
    gone.Flag
    gone.Logger `gone:"*"`
}

func (p *ThirdBusiness1Provider) Provide(tagConf string) (*ThirdBusiness1, error) {
    p.Infof("tagConf->%s", tagConf)
    return &ThirdBusiness1{Name: "ThirdBusiness1"}, nil
}

// 无参数 Provider 实现
type ThirdBusiness2Provider struct {
    gone.Flag
}

func (p *ThirdBusiness2Provider) Provide() (*ThirdBusiness2, error) {
    return &ThirdBusiness2{Name: "ThirdBusiness2"}, nil
}

type ThirdBusinessUser struct {
    gone.Flag
    thirdBusiness1 *ThirdBusiness1 `gone:"*,AGI"`
    thirdBusiness2 *ThirdBusiness2 `gone:"*"`
}

func main() {
    gone.
        Load(&ThirdBusinessUser{}).
        Load(&ThirdBusiness1Provider{}).
        Load(&ThirdBusiness2Provider{}).
        Run(func(user *ThirdBusinessUser, log gone.Logger) {
            log.Infof("user.thirdBusiness1.name->%s", user.thirdBusiness1.Name)
            log.Infof("user.thirdBusiness2.name->%s", user.thirdBusiness2.Name)
        })
}
```

运行结果：
```log
2025/03/11 10:03:22 tagConf->AGI
2025/03/11 10:03:22 user.thirdBusiness1.name->ThirdBusiness1
2025/03/11 10:03:22 user.thirdBusiness2.name->ThirdBusiness2
```

### 2.2 按名字注入

当存在相同类型的多个 Provider 时，可以通过名字进行区分。字段上的 `gone` 标签中指定 Provider 名称，要求 Provider 返回的对象类型与字段类型兼容。

**示例代码：**
```go
package main

import "github.com/gone-io/gone/v2"

type ThirdBusiness struct {
    Name string
}

type xProvider struct {
    gone.Flag
}

func (p *xProvider) GonerName() string {
    return "x-business-provider"
}

func (p *xProvider) Provide(tagConf string) (*ThirdBusiness, error) {
    return &ThirdBusiness{Name: "x-" + tagConf}, nil
}

type yProvider struct {
    gone.Flag
}

func (p *yProvider) GonerName() string {
    return "y-business-provider"
}

func (p *yProvider) Provide() (*ThirdBusiness, error) {
    return &ThirdBusiness{Name: "y"}, nil
}

type ThirdBusinessUser struct {
    gone.Flag
    x *ThirdBusiness `gone:"x-business-provider,extend"`
    y *ThirdBusiness `gone:"y-business-provider"`
}

func main() {
    gone.
        Load(&ThirdBusinessUser{}).
        Load(&xProvider{}, gone.OnlyForName()).
        Load(&yProvider{}, gone.OnlyForName()).
        Run(func(user *ThirdBusinessUser, log gone.Logger) {
            log.Infof("user.x.name->%s", user.x.Name)
            log.Infof("user.y.name->%s", user.y.Name)
        })
}
```

在注册 Provider 时有两种指定名字的方法：
1. **实现 `GonerName()` 方法** —— 返回 Provider 的名称。
2. **在 `Load` 时使用 `gone.Name("...")` 选项** —— 显式指定 Provider 名称。

> **提示：** 当使用相同类型提供多个 Provider 时，需要使用 `gone.OnlyForName()` 选项，否则框架会报错；而相同名字的 Provider 只能存在一个，否则也会报错。

### 2.3 基于名字的多类型 Provider

在某些场景下（如配置注入），希望通过一个 Provider 提供多种类型的对象，此时可以定义 `NamedProvider` 接口：
```go
type NamedProvider interface {
    NamedGoner
    Provide(tagConf string, t reflect.Type) (any, error)
}
```
其中：
- `tagConf` 参数为 `gone` 标签中第一个逗号后面的扩展配置部分。
- `t` 参数表示需要注入字段的类型。

这种设计允许一个 Provider 根据字段类型返回对应的实例，实现多类型注入。

---

## 3. “星号” Provider：`*`

当 `gone` 标签中的名称为 `*` 或省略时，表示按类型注入。这实际上是框架中预定义的一个 `NamedProvider`，其名称就是 `*`。  
其工作逻辑为：
- 根据需要注入字段的类型，自动查找并调用合适的 Provider 来提供值。

---

## 总结

Gone V2 通过 Provider 机制实现了灵活的依赖注入，支持：

- **按类型注入**：直接根据字段类型寻找合适的 Provider。
- **按名字注入**：在存在多个相同类型 Provider 时，通过名称进行区分。
- **多类型 Provider**：一个 Provider 可根据字段类型返回不同的对象，适用于配置注入等场景。
- **“星号” Provider**：简化按类型注入的处理流程。

这种设计不仅降低了第三方对象与框架耦合度，也大大提高了依赖注入的灵活性和扩展性。

---