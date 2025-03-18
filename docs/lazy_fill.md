# Gone框架中的延迟依赖注入：LazyFill()与option:"lazy"标签

在Gone框架中，循环依赖并不是一个常见问题，但在某些场景下，当多个组件之间存在相互依赖关系时可能会出现。为了解决这个问题，Gone框架提供了两种延迟依赖注入的机制：`LazyFill()`选项和`option:"lazy"`标签。本文将详细介绍这两种机制的工作原理、使用场景以及它们之间的异同点。

## 循环依赖问题回顾

在深入了解延迟依赖注入机制之前，让我们先回顾一下Gone框架中的循环依赖问题。

Gone框架的初始化流程主要分为两个阶段：
1. **fillAction（字段填充）**：框架将依赖注入到组件的字段中
2. **initAction（组件初始化）**：框架调用组件的`Init()`方法进行初始化

当两个或多个组件之间存在相互依赖，并且它们都实现了`Init()`方法时，就会形成一个无法解决的循环：
- A的字段填充依赖B的初始化
- B的初始化依赖B的字段填充
- B的字段填充依赖A的初始化
- A的初始化依赖A的字段填充

这种情况下，框架无法确定初始化顺序，因此会抛出循环依赖的错误。

## LazyFill()选项

### 工作原理

`LazyFill()`是一个加载选项，用于标记一个Goner为延迟填充。当使用这个选项时，被标记的Goner的装配过程（fillAction）会被延迟到其他组件装配后进行。

在Gone框架的内部实现中，`LazyFill()`选项会设置coffin（Goner的包装器）的`lazyFill`属性为true：

```go
func LazyFill() Option {
    return option{
        apply: func(c *coffin) error {
            c.lazyFill = true
            return nil
        },
    }
}
```

当框架收集依赖关系时，如果一个Goner被标记为`lazyFill`，它的fillAction不会被添加到其他组件的依赖列表中：

```go
func (s *Core) getGonerDeps(co *coffin) (fillDependencies, initDependencies []dependency, err error) {
    fillDependencies, err = s.getGonerFillDeps(co)
    if !co.lazyFill {
        initDependencies = append(initDependencies, dependency{
            coffin: co,
            action: fillAction,
        })
    }
    return
}
```

### 使用示例

以下是使用`LazyFill()`选项解决循环依赖的示例：

```go
type depA5 struct {
    gone.Flag
    dep *depB5 `gone:"*"`
}

func (d *depA5) Init() {
    if d.dep != nil {
        panic("depB4.dep should be nil")
    }
}

type depB5 struct {
    gone.Flag
    dep *depA5 `gone:"*"`
}

func (d *depB5) Init() {
    if d.dep == nil {
        panic("depB4.dep should not be nil")
    }
}

func TestCircularDependency5(t *testing.T) {
    gone.
        NewApp().
        Load(&depB5{}).
        Load(&depA5{}, gone.LazyFill()). // 使用LazyFill()选项
        Run(func(a4 *depA5, b4 *depB5) {
            if a4.dep == nil {
                t.Error("a4.dep should be not nil")
            }
            if b4.dep == nil {
                t.Error("b4.dep should be not nil")
            }
        })
}
```

在这个例子中，`depA5`被标记为延迟填充，这意味着它的装配过程会被延迟，从而打破了循环依赖。

### 注意事项

使用`LazyFill()`选项时，需要注意以下几点：

1. 被延迟的Goner在`Init()`、`Provide()`、`Inject()`等方法中，无法使用依赖注入的字段，因为这些方法可能在字段填充之前被调用。
2. `LazyFill()`是一个全局选项，会影响整个Goner的装配过程。

## option:"lazy"标签

### 工作原理

`option:"lazy"`是一个字段标签，用于标记特定字段为延迟注入。当一个字段被标记为lazy时，该字段不会在依赖收集阶段被考虑，从而避免形成循环依赖。

在Gone框架的内部实现中，`isLazyField()`函数用于检查一个字段是否被标记为lazy：

```go
func isLazyField(filed *reflect.StructField) bool {
    return filedHasOption(filed, optionTag, lazy)
}
```

当框架收集依赖关系时，会跳过被标记为lazy的字段：

```go
func (s *Core) getGonerFillDeps(co *coffin) (fillDependencies []dependency, err error) {
    // ...
    for i := 0; i < elem.NumField(); i++ {
        field := elem.Field(i)

        if isLazyField(&field) {
            continue // 跳过lazy字段
        }
        // 处理其他字段...
    }
    // ...
}
```

### 使用示例

以下是使用`option:"lazy"`标签解决循环依赖的示例：

```go
type depA4 struct {
    gone.Flag
    dep *depB4 `gone:"*"`
}

func (d *depA4) Init() {
    if d.dep == nil {
        panic("depB4.dep should not be nil")
    }
}

type depB4 struct {
    gone.Flag
    dep *depA4 `gone:"*" option:"lazy"` // 使用option:"lazy"标签
}

func (d *depB4) Init() {
    if d.dep.dep != nil {
        panic("depB4.dep should be nil")
    }
}

func TestCircularDependency4(t *testing.T) {
    gone.
        NewApp().
        Load(&depA4{}).
        Load(&depB4{}).
        Run(func(a4 *depA4, b4 *depB4) {
            if a4.dep == nil {
                t.Error("a4.dep should be not nil")
            }
            if b4.dep == nil {
                t.Error("b4.dep should be not nil")
            }
        })
}
```

在这个例子中，`depB4`的`dep`字段被标记为lazy，这意味着在依赖收集阶段，这个字段不会被考虑，从而打破了循环依赖。

### 注意事项

使用`option:"lazy"`标签时，需要注意以下几点：

1. 被标记为lazy的字段不能在`Init()`、`Provide()`、`Inject()`等方法中使用，因为这些方法可能在字段填充之前被调用。
2. `option:"lazy"`是一个字段级别的选项，只影响特定字段的装配过程。

## LazyFill()与option:"lazy"的异同点

### 相同点

1. **目的相同**：两者都是为了解决循环依赖问题。
2. **原理相似**：都是通过延迟依赖注入来打破循环依赖。
3. **使用限制**：被延迟的依赖都不能在`Init()`、`Provide()`、`Inject()`等方法中使用。

### 不同点

1. **作用范围**：
   - `LazyFill()`是一个全局选项，影响整个Goner的装配过程。
   - `option:"lazy"`是一个字段级别的选项，只影响特定字段的装配过程。

2. **使用方式**：
   - `LazyFill()`在加载Goner时使用：`Load(&myGoner{}, gone.LazyFill())`
   - `option:"lazy"`在定义字段时使用：`dep *AnotherGoner `gone:"*" option:"lazy"`

3. **灵活性**：
   - `option:"lazy"`更加灵活，可以精确控制哪些字段需要延迟注入。
   - `LazyFill()`更加简单，一次性解决整个Goner的循环依赖问题。

## 选择指南

在实际应用中，如何选择使用`LazyFill()`还是`option:"lazy"`？以下是一些建议：

1. **当整个组件都需要延迟装配时**，使用`LazyFill()`更加简单直接。
2. **当只有特定字段需要延迟注入时**，使用`option:"lazy"`更加精确。
3. **当需要精细控制依赖关系时**，可以组合使用两种机制。

## 最佳实践

1. **优先考虑重构组件设计**：在使用延迟依赖注入之前，应该先考虑是否可以通过重构组件设计来消除循环依赖。
2. **明确依赖关系**：在使用延迟依赖注入时，应该明确了解组件之间的依赖关系，避免引入新的问题。
3. **谨慎使用Init方法**：如果可能，尽量减少使用`Init()`方法，或者确保`Init()`方法不依赖于被延迟注入的字段。
4. **文档化延迟依赖**：在代码中明确标注哪些依赖是延迟注入的，以便其他开发者理解代码。

## 总结

Gone框架提供了两种延迟依赖注入的机制：`LazyFill()`选项和`option:"lazy"`标签，它们都可以有效解决循环依赖问题。在实际应用中，应该根据具体需求选择合适的机制，并遵循最佳实践，以确保代码的可维护性和可靠性。

通过合理使用这两种机制，可以在保持组件间依赖关系的同时，避免循环依赖带来的问题，从而构建更加健壮的应用程序。