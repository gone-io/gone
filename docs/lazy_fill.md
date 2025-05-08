<p>
   English&nbsp ｜&nbsp <a href="lazy_fill_CN.md">中文</a>
</p>

# Lazy Dependency Injection in Gone Framework: LazyFill() and option:"lazy" Tag

- [Lazy Dependency Injection in Gone Framework: LazyFill() and option:"lazy" Tag](#lazy-dependency-injection-in-gone-framework-lazyfill-and-optionlazy-tag)
  - [Revisiting the Circular Dependency Problem](#revisiting-the-circular-dependency-problem)
  - [The LazyFill() Option](#the-lazyfill-option)
    - [Working Principle](#working-principle)
    - [Usage Example](#usage-example)
    - [Considerations](#considerations)
  - [The option:"lazy" Tag](#the-optionlazy-tag)
    - [Working Principle](#working-principle-1)
    - [Usage Example](#usage-example-1)
    - [Considerations](#considerations-1)
  - [Similarities and Differences Between LazyFill() and option:"lazy"](#similarities-and-differences-between-lazyfill-and-optionlazy)
    - [Similarities](#similarities)
    - [Differences](#differences)
  - [Selection Guide](#selection-guide)
  - [Best Practices](#best-practices)
  - [Summary](#summary)


In the Gone framework, circular dependencies are not a common problem, but they may occur in certain scenarios when multiple components have interdependent relationships. To solve this problem, the Gone framework provides two mechanisms for lazy dependency injection: the `LazyFill()` option and the `option:"lazy"` tag. This article will detail the working principles, use cases, and differences between these two mechanisms.

## Revisiting the Circular Dependency Problem

Before diving into lazy dependency injection mechanisms, let's review the circular dependency problem in the Gone framework.

The initialization process of the Gone framework is mainly divided into two phases:
1. **fillAction (field filling)**: The framework injects dependencies into component fields
2. **initAction (component initialization)**: The framework calls the component's `Init()` method for initialization

When two or more components have mutual dependencies and they all implement the `Init()` method, an unsolvable cycle is formed:
- A's field filling depends on B's initialization
- B's initialization depends on B's field filling
- B's field filling depends on A's initialization
- A's initialization depends on A's field filling

In this situation, the framework cannot determine the initialization order and will throw a circular dependency error.

## The LazyFill() Option

### Working Principle

`LazyFill()` is a loading option used to mark a Goner for lazy filling. When this option is used, the assembly process (fillAction) of the marked Goner will be delayed until after other components are assembled.

In the internal implementation of the Gone framework, the `LazyFill()` option sets the `lazyFill` property of the coffin (Goner's wrapper) to true:

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

When the framework collects dependency relationships, if a Goner is marked as `lazyFill`, its fillAction will not be added to the dependency list of other components:

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

### Usage Example

Here is an example of using the `LazyFill()` option to solve circular dependencies:

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
        Load(&depA5{}, gone.LazyFill()). // Using LazyFill() option
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

In this example, `depA5` is marked for lazy filling, which means its assembly process will be delayed, thus breaking the circular dependency.

### Considerations

When using the `LazyFill()` option, note the following:

1. The delayed Goner cannot use dependency-injected fields in methods like `Init()`, `Provide()`, `Inject()`[1], as these methods might be called before field filling.
2. `LazyFill()` is a global option that affects the entire assembly process of the Goner.

## The option:"lazy" Tag

### Working Principle

`option:"lazy"` is a field tag used to mark specific fields for lazy injection. When a field is marked as lazy, it won't be considered during the dependency collection phase, thus avoiding the formation of circular dependencies.

In the internal implementation of the Gone framework, the `isLazyField()` function is used to check if a field is marked as lazy:

```go
func isLazyField(filed *reflect.StructField) bool {
    return filedHasOption(filed, optionTag, lazy)
}
```

When the framework collects dependency relationships, it skips fields marked as lazy:

```go
func (s *Core) getGonerFillDeps(co *coffin) (fillDependencies []dependency, err error) {
    // ...
    for i := 0; i < elem.NumField(); i++ {
        field := elem.Field(i)

        if isLazyField(&field) {
            continue // Skip lazy fields
        }
        // Process other fields...
    }
    // ...
}
```

### Usage Example

Here is an example of using the `option:"lazy"` tag to solve circular dependencies:

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
    dep *depA4 `gone:"*" option:"lazy"` // Using option:"lazy" tag
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

In this example, the `dep` field of `depB4` is marked as lazy, which means it won't be considered during the dependency collection phase, thus breaking the circular dependency.

### Considerations

When using the `option:"lazy"` tag, note the following:

1. Fields marked as lazy cannot be used in methods like `Init()`, `Provide()`, `Inject()`[1], as these methods might be called before field filling.
2. `option:"lazy"` is a field-level option that only affects the assembly process of specific fields.

## Similarities and Differences Between LazyFill() and option:"lazy"

### Similarities

1. **Same Purpose**: Both are designed to solve circular dependency problems.
2. **Similar Principle**: Both break circular dependencies by delaying dependency injection.
3. **Usage Restrictions**: Delayed dependencies cannot be used in methods like `Init()`, `Provide()`, `Inject()`[1].

### Differences

1. **Scope**:
    - `LazyFill()` is a global option that affects the entire assembly process of a Goner.
    - `option:"lazy"` is a field-level option that only affects the assembly process of specific fields.

2. **Usage Method**:
    - `LazyFill()` is used when loading a Goner: `Load(&myGoner{}, gone.LazyFill())`
    - `option:"lazy"` is used when defining a field: `dep *AnotherGoner `gone:"*" option:"lazy"`

3. **Flexibility**:
    - `option:"lazy"` is more flexible, allowing precise control over which fields need lazy injection.
    - `LazyFill()` is simpler, solving the circular dependency problem for an entire Goner at once.

## Selection Guide

In practical applications, how to choose between `LazyFill()` and `option:"lazy"`? Here are some suggestions:

1. **When an entire component needs delayed assembly**, using `LazyFill()` is simpler and more direct.
2. **When only specific fields need lazy injection**, using `option:"lazy"` is more precise.
3. **When fine-grained control of dependencies is needed**, both mechanisms can be combined.

## Best Practices

1. **Prioritize Refactoring Component Design**: Before using lazy dependency injection, consider whether the circular dependency can be eliminated by refactoring the component design.
2. **Clarify Dependency Relationships**: When using lazy dependency injection, clearly understand the dependencies between components to avoid introducing new problems.
3. **Use the Init Method Cautiously**: If possible, minimize the use of the `Init()` method, or ensure that the `Init()` method does not depend on fields that are lazily injected.
4. **Document Lazy Dependencies**: Clearly annotate which dependencies are lazily injected in the code for other developers to understand.

## Summary

The Gone framework provides two mechanisms for lazy dependency injection: the `LazyFill()` option and the `option:"lazy"` tag, both of which can effectively solve circular dependency problems. In practical applications, choose the appropriate mechanism based on specific needs and follow best practices to ensure code maintainability and reliability.

By using these two mechanisms properly, circular dependency problems can be avoided while maintaining component relationships, resulting in more robust applications.

Note:
1. Methods like `Init()`, `Provide()`, `Inject()` include:
    - `Init()` Init method without return value
    - `Init() error` Init method with return value
    - `Provide(tagConf string) (T, error)` Provide method with tagConf parameter
    - `Provide() (T, error)` Provide method without parameters
    - `Provide(tagConf string, t reflect.Type) (any, error)` Provide method that provides values by type
    - `Inject(tagConf string, field reflect.StructField, fieldValue reflect.Value) error` Inject method that can inject values into fields