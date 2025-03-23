# 依赖注入介绍

- [依赖注入介绍](#依赖注入介绍)
  - [什么是 Goner？](#什么是-goner)
    - [`gone` 标签的格式](#gone-标签的格式)
    - [代码示例](#代码示例)
  - [如何将 Goner 注册到 Gone 框架？](#如何将-goner-注册到-gone-框架)
    - [方式一：单个注册](#方式一单个注册)
    - [方式二：批量注册](#方式二批量注册)
  - [依赖注入的执行时机](#依赖注入的执行时机)
  - [手动完成依赖注入](#手动完成依赖注入)
    - [方式一：结构体注入（StructInjector）](#方式一结构体注入structinjector)
    - [方式二：函数参数注入（FuncInjector）](#方式二函数参数注入funcinjector)
  - [总结](#总结)


在使用 Gone 框架时，你可能会好奇它是如何进行依赖注入的。本文将带你深入了解 Gone 的注入机制，并通过示例让你轻松掌握它的用法。

---

## 什么是 Goner？

在 Gone 框架中，被注入的对象称为 **Goner**。但要让对象成为 Goner，它必须满足以下两个条件：

1. **对象必须嵌入 `gone.Flag`** —— 这个 Flag 让 Gone 识别出它是一个 Goner。
2. **需要注入的字段必须标记 `gone` 标签** —— 只有标记了 `gone` 标签的字段才会被注入。

### `gone` 标签的格式

```go
gone:"${name},${extend}"
```

- **`${name}`**：表示 Goner 的名字。如果是 `*` 或者省略，表示按类型自动注入。
- **`${extend}`**（可选）：扩展选项，会传递给 Provider 的 `Provide` 方法，在[《Gone V2 Provider 机制介绍》](./provider.md)中有更详细的介绍。

### 代码示例

```go
type Dep struct {
    gone.Flag
}

type Dep2 struct {
    gone.Flag
}

type UseDep struct {
    gone.Flag
    dep  *Dep  `gone:"*"`   // 自动按类型注入
    Dep2 *Dep2 `gone:"dep2"` // 按名称 "dep2" 注入
}
```

:::tip
**字段可以是私有的！** 这样符合**开放封闭原则**，更安全，也更符合封装设计。
:::

---

## 如何将 Goner 注册到 Gone 框架？

注册 Goner 其实非常简单，通常有两种方式：

### 方式一：单个注册

```go
gone.Load(&UseDep{})
```

### 方式二：批量注册

如果你需要一次性注册多个 Goner，可以这样写：

```go
gone.Loads(func(l gone.Loader) error {
    _ = l.Load(&UseDep{})
    _ = l.Load(&Dep{})
    _ = l.Load(&Dep2{}, gone.Name("dep2"))
    return nil
})
```

在注册时，还可以传递多个扩展参数，例如 `gone.Name()` 指定 Goner 的名称。

---

## 依赖注入的执行时机

当框架启动后，所有注册的对象都会自动执行依赖注入。如果有字段的 `gone` 标签找不到对应的依赖，框架会直接报错，确保依赖完整性。

---

## 手动完成依赖注入

虽然 Gone 可以自动注入，但有时我们希望**手动控制**依赖注入。框架提供了两种手动注入方式：

1. **`gone.StructInjector`** —— 用于结构体字段注入。
2. **`gone.FuncInjector`** —— 用于函数参数注入。

### 方式一：结构体注入（StructInjector）

假设我们有一个 `Business` 结构体，它依赖 `Dep`，但 `Dep` 不是一开始就存在的，需要在运行时注入进去。

```go
package main

import "github.com/gone-io/gone/v2"

type Business struct {
    gone.Flag
    structInjector gone.StructInjector `gone:"*"`
}

type Dep struct {
    gone.Flag
    Name string
}

func (b *Business) BusinessProcess() {
    type User struct {
        Dep *Dep `gone:"*"`
    }
    
    var user User
    err := b.structInjector.InjectStruct(&user)
    if err != nil {
        panic(err)
    }
    println("user.Dep.Name->", user.Dep.Name)
}

func main() {
    gone.
        Load(&Business{}).
        Load(&Dep{Name: "dep"}).
        Run(func(b *Business) {
            b.BusinessProcess()
        })
}
```

### 方式二：函数参数注入（FuncInjector）

如果你有一个函数 `needInjectedFunc` 需要依赖 `Dep`，你可以用 `gone.FuncInjector` 自动注入它的参数。

```go
package main

import "github.com/gone-io/gone/v2"

type Business struct {
    gone.Flag
    funcInjector gone.FuncInjector `gone:"*"`
}

type Dep struct {
    gone.Flag
    Name string
}

func (b *Business) BusinessProcess() {
    needInjectedFunc := func(dep *Dep) {
        println("dep.name->", dep.Name)
    }

    wrapFunc, err := b.funcInjector.InjectWrapFunc(needInjectedFunc, nil, nil)
    if err != nil {
        panic(err)
    }
    _ = wrapFunc()
}

func main() {
    gone.
        Load(&Business{}).
        Load(&Dep{Name: "dep"}).
        Run(func(b *Business) {
            b.BusinessProcess()
        })
}
```

---

## 总结

- **Gone 框架使用 `gone.Flag` 和 `gone` 标签进行自动依赖注入。**
- **Goner 可以按类型 (`*`) 或者按名称注入 (`gone:"dep2"`)。**
- **对象必须先注册到框架，才能参与依赖注入。**
- **如果自动注入不够用，还可以使用 `gone.StructInjector` 和 `gone.FuncInjector` 进行手动注入。**

掌握这些基本概念后，你就可以在项目中高效地使用 Gone 框架的依赖注入机制了！🎉

