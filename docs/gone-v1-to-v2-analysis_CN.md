<p>
    <a href="gone-v1-to-v2-analysis.md">English</a>&nbsp ｜&nbsp 中文
</p>

# Gone 从 v1 到 v2 的更新分析

- [Gone 从 v1 到 v2 的更新分析](#gone-从-v1-到-v2-的更新分析)
  - [1. 概念简化与术语变更](#1-概念简化与术语变更)
  - [2. 接口重新设计](#2-接口重新设计)
    - [2.1 组件定义的简化](#21-组件定义的简化)
    - [2.2 组件加载方式的统一](#22-组件加载方式的统一)
    - [2.3 生命周期方法的优化](#23-生命周期方法的优化)
  - [3. 依赖注入逻辑重写](#3-依赖注入逻辑重写)
    - [3.1 注入标签的简化](#31-注入标签的简化)
    - [3.2 依赖注入查找流程优化](#32-依赖注入查找流程优化)
  - [4. Provider 机制的引入](#4-provider-机制的引入)
    - [4.1 泛型 Provider 接口](#41-泛型-provider-接口)
    - [4.2 NamedProvider 接口](#42-namedprovider-接口)
    - [4.3 NoneParamProvider 接口](#43-noneparamprovider-接口)
  - [5. 多实例支持](#5-多实例支持)
  - [6. 动态组件获取](#6-动态组件获取)
  - [7. 函数参数注入](#7-函数参数注入)
  - [8. 仓库结构优化](#8-仓库结构优化)
  - [9. 迁移指南](#9-迁移指南)
  - [10. 总结](#10-总结)


Gone 框架在 v2 版本中进行了全面的更新和改进，主要目标是简化框架概念，提高易用性和性能。本文档将详细分析 v1 到 v2 的主要变化。

## 1. 概念简化与术语变更

v1 版本中，Gone 框架使用了大量宗教性的概念和术语来描述框架的各个部分，这些术语在 v2 版本中被替换为更加直观和技术性的术语：

| v1 版本术语 | v2 版本术语 | 描述 |
|------------|------------|------|
| Heaven | Application | 应用程序实例，负责管理组件的生命周期 |
| Cemetery | Core | 框架核心，负责组件的注册和管理 |
| Priest | Loader | 组件加载器，负责将组件加载到框架中 |
| Goner | Goner | 保留，但定义更加明确，指嵌入了 `gone.Flag` 的结构体指针 |
| Prophet | 移除 | v2 版本使用更清晰的生命周期方法替代 |
| Angel | 移除 | v2 版本使用更清晰的生命周期方法替代 |
| Vampire | Provider | 转变为更直观的 Provider 机制 |
| Tomb | 移除 | 简化了相关概念 |

这种术语变更使框架更加专业和易于理解，降低了学习曲线，使开发者能够更快地上手使用框架。

## 2. 接口重新设计

v2 版本重新设计了框架的接口，减少了内部方法的暴露，使接口更加清晰和易于使用：

### 2.1 组件定义的简化

v1 版本中，组件需要嵌入 `gone.GonerFlag`，而在 v2 版本中，组件只需要嵌入 `gone.Flag`：

```go
// v1 版本
type Component struct {
    gone.GonerFlag
}

// v2 版本
type Component struct {
    gone.Flag
}
```

### 2.2 组件加载方式的统一

v2 版本提供了更加一致和灵活的组件加载方式：

```go
// v1 版本
func Priest(cemetery gone.Cemetery) error {
    cemetery.Bury(&Component{}, "component-id")
    return nil
}

// v2 版本
gone.Load(&Component{})                        // 直接加载
gone.Load(&Component{}, gone.Name("component")) // 命名加载
gone.Load(&Component{}).Load(&Component2{})    // 链式加载
```

### 2.3 生命周期方法的优化

v2 版本优化了组件的生命周期管理，使其更加清晰和可预测：

```go
// v1 版本中的 Prophet 和 Angel
type Prophet interface {
    AfterRevive(gone.Cemetery, gone.Tomb) gone.ReviveAfterError
}

type Angel interface {
    Start(gone.Cemetery) error
    Stop(gone.Cemetery) error
}

// v2 版本中的生命周期方法
type Initer interface {
    Init() error
}

type Starter interface {
    Start() error
}

type Stopper interface {
    Stop() error
}
```

## 3. 依赖注入逻辑重写

v2 版本重写了依赖注入的实现逻辑，使其更加灵活和强大：

### 3.1 注入标签的简化

```go
// v1 版本
type Service struct {
    gone.GonerFlag
    Dep *Dependency `gone:"dependency-id"`
}

// v2 版本
type Service struct {
    gone.Flag
    Dep *Dependency `gone:"dependency"` // 基于名称注入
    Dep2 *Dependency `gone:"*"`         // 基于类型注入
}
```

### 3.2 依赖注入查找流程优化

v2 版本明确了依赖注入时查找组件的优先级和流程，使注入过程更加可预测：

1. 首先根据标签中指定的名称查找
2. 如果未找到，则根据字段类型查找
3. 如果存在多个相同类型的组件，优先选择默认实现（通过 `IsDefault()` 选项设置）

## 4. Provider 机制的引入

v2 版本引入了全新的 Provider 机制，替代了 v1 版本中的 Vampire 概念：

### 4.1 泛型 Provider 接口

```go
type Provider[T any] interface {
    Goner
    Provide(tagConf string) (T, error)
}
```

### 4.2 NamedProvider 接口

```go
type NamedProvider interface {
    NamedGoner
    Provide(tagConf string, t reflect.Type) (any, error)
}
```

### 4.3 NoneParamProvider 接口

```go
type NoneParamProvider[T any] interface {
    Goner
    Provide() T
}
```

Provider 机制使组件能够动态创建和提供其他组件，大大增强了框架的灵活性和扩展性。

## 5. 多实例支持

v2 版本增强了对多实例的支持，允许在同一个应用程序中创建多个 Gone 框架实例：

```go
// 创建多个 Gone 框架实例
app1 := gone.NewApp()
app2 := gone.NewApp()

// 每个实例可以独立加载组件和运行
app1.Load(&Component1{})
app2.Load(&Component2{})

app1.Run()
app2.Run()
```

## 6. 动态组件获取

v2 版本提供了更灵活的动态组件获取方式：

```go
type GonerKeeper interface {
    GetGonerByName(name string) any
    GetGonerByType(t reflect.Type) any
}
```

## 7. 函数参数注入

v2 版本增强了函数参数注入功能：

```go
type FuncInjector interface {
    InjectWrapFunc(fn interface{}, args []interface{}, kwargs map[string]interface{}) (func() error, error)
}
```

## 8. 仓库结构优化

v2 版本进行了重要的仓库结构调整，将 `github.com/gone-io/gone/goner` 独立成了单独的仓库进行管理，而 `github.com/gone-io/gone` 仓库则专注于管理 Gone 依赖注入的核心代码：

```
// v1 版本
github.com/gone-io/gone        // 包含所有 Gone 框架代码
github.com/gone-io/gone/goner  // 作为主仓库的子目录

// v2 版本
github.com/gone-io/gone       // 只包含依赖注入核心代码
github.com/gone-io/goner      // 独立仓库，管理 Goner 相关代码
```

这种仓库结构的优化带来了以下好处：

1. **更清晰的模块边界**：通过将 Goner 相关代码独立出来，使框架的模块边界更加清晰，每个仓库都有明确的职责和功能范围。

2. **更灵活的版本管理**：独立的仓库可以有独立的版本发布周期，使 Goner 模块可以根据自身需求进行迭代和更新，而不必与主框架同步。

3. **更好的代码复用性**：独立的 Goner 仓库可以被其他项目更方便地引用和复用，不必引入整个 Gone 框架。

4. **更专注的维护职责**：团队可以根据专长分工维护不同的仓库，提高开发效率和代码质量。

5. **降低依赖复杂度**：使用者可以根据实际需求选择性地引入所需模块，减少不必要的依赖。

这种仓库结构的调整反映了 Gone 框架在架构设计上的持续优化，使框架更加模块化和可维护。

## 9. 迁移指南

从 v1 迁移到 v2 版本需要注意以下几点：

1. **更新导入路径**：使用 `github.com/gone-io/gone/v2` 替代 `github.com/gone-io/gone`

2. **调整组件定义**：确保所有组件都嵌入了 `gone.Flag`

3. **使用新的加载方式**：采用 v2 版本提供的组件加载方式

4. **适应新的 Provider 机制**：如果使用了自定义 Provider，需要调整为 v2 版本的 Provider 接口

5. **检查生命周期方法**：确保生命周期方法符合 v2 版本的规范

## 10. 总结

Gone v2 版本通过以下几个方面的改进，使框架更加易用、灵活和强大：

1. **概念简化**：移除宗教性术语，使用更直观的技术术语
2. **接口重新设计**：减少内部方法的暴露，使接口更加清晰
3. **组件加载机制改进**：提供更加一致和灵活的组件加载方式
4. **Provider 机制引入**：替代 Vampire 概念，提供更强大的组件创建和提供能力
5. **生命周期管理优化**：使组件的生命周期更加清晰和可预测
6. **多实例支持**：允许在同一个应用程序中创建多个 Gone 框架实例
7. **动态组件获取**：提供更灵活的动态组件获取方式
8. **函数参数注入**：增强函数参数注入功能
9. **仓库结构优化**：将 Goner 独立成单独仓库，使框架更加模块化和可维护

这些改进使 Gone 框架更适合构建复杂的应用程序，特别是微服务架构的应用。