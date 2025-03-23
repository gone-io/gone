# Gone Framework Update Analysis from v1 to v2

- [Gone Framework Update Analysis from v1 to v2](#gone-framework-update-analysis-from-v1-to-v2)
  - [1. Concept Simplification and Terminology Changes](#1-concept-simplification-and-terminology-changes)
  - [2. Interface Redesign](#2-interface-redesign)
    - [2.1 Component Definition Simplification](#21-component-definition-simplification)
    - [2.2 Unified Component Loading](#22-unified-component-loading)
    - [2.3 Lifecycle Method Optimization](#23-lifecycle-method-optimization)
  - [3. Dependency Injection Logic Rewrite](#3-dependency-injection-logic-rewrite)
    - [3.1 Injection Tag Simplification](#31-injection-tag-simplification)
    - [3.2 Dependency Injection Lookup Process Optimization](#32-dependency-injection-lookup-process-optimization)
  - [4. Introduction of Provider Mechanism](#4-introduction-of-provider-mechanism)
    - [4.1 Generic Provider Interface](#41-generic-provider-interface)
    - [4.2 NamedProvider Interface](#42-namedprovider-interface)
    - [4.3 NoneParamProvider Interface](#43-noneparamprovider-interface)
  - [5. Multiple Instance Support](#5-multiple-instance-support)
  - [6. Dynamic Component Retrieval](#6-dynamic-component-retrieval)
  - [7. Function Parameter Injection](#7-function-parameter-injection)
  - [8. Repository Structure Optimization](#8-repository-structure-optimization)
  - [9. Migration Guide](#9-migration-guide)
  - [10. Summary](#10-summary)


The Gone framework has undergone a comprehensive update and improvement in v2, with the main goals of simplifying framework concepts, enhancing usability, and improving performance. This document will analyze in detail the major changes from v1 to v2.

## 1. Concept Simplification and Terminology Changes

In v1, the Gone framework used many religious concepts and terms to describe different parts of the framework. These terms have been replaced in v2 with more intuitive and technical terminology:

| v1 Term | v2 Term | Description |
|------------|------------|------|
| Heaven | Application | Application instance, responsible for managing component lifecycle |
| Cemetery | Core | Framework core, responsible for component registration and management |
| Priest | Loader | Component loader, responsible for loading components into the framework |
| Goner | Goner | Retained, but with a more precise definition - refers to a structure pointer embedded with `gone.Flag` |
| Prophet | Removed | v2 uses clearer lifecycle methods as replacements |
| Angel | Removed | v2 uses clearer lifecycle methods as replacements |
| Vampire | Provider | Transformed into a more intuitive Provider mechanism |
| Tomb | Removed | Related concepts have been simplified |

These terminology changes make the framework more professional and easier to understand, lowering the learning curve and enabling developers to start using the framework more quickly.

## 2. Interface Redesign

v2 has redesigned the framework interfaces, reducing exposure of internal methods and making interfaces clearer and easier to use:

### 2.1 Component Definition Simplification

In v1, components needed to embed `gone.GonerFlag`, while in v2, components only need to embed `gone.Flag`:

```go
// v1 version
type Component struct {
    gone.GonerFlag
}

// v2 version
type Component struct {
    gone.Flag
}
```

### 2.2 Unified Component Loading

v2 provides a more consistent and flexible approach to component loading:

```go
// v1 version
func Priest(cemetery gone.Cemetery) error {
    cemetery.Bury(&Component{}, "component-id")
    return nil
}

// v2 version
gone.Load(&Component{})                        // Direct loading
gone.Load(&Component{}, gone.Name("component")) // Named loading
gone.Load(&Component{}).Load(&Component2{})    // Chain loading
```

### 2.3 Lifecycle Method Optimization

v2 has optimized component lifecycle management, making it clearer and more predictable:

```go
// v1 version Prophet and Angel
type Prophet interface {
    AfterRevive(gone.Cemetery, gone.Tomb) gone.ReviveAfterError
}

type Angel interface {
    Start(gone.Cemetery) error
    Stop(gone.Cemetery) error
}

// v2 version lifecycle methods
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

## 3. Dependency Injection Logic Rewrite

v2 has rewritten the dependency injection implementation logic, making it more flexible and powerful:

### 3.1 Injection Tag Simplification

```go
// v1 version
type Service struct {
    gone.GonerFlag
    Dep *Dependency `gone:"dependency-id"`
}

// v2 version
type Service struct {
    gone.Flag
    Dep *Dependency `gone:"dependency"` // Name-based injection
    Dep2 *Dependency `gone:"*"`         // Type-based injection
}
```

### 3.2 Dependency Injection Lookup Process Optimization

v2 clarifies the priority and process for finding components during dependency injection, making the injection process more predictable:

1. First search based on the name specified in the tag
2. If not found, search based on field type
3. If multiple components of the same type exist, prioritize the default implementation (set via the `IsDefault()` option)

## 4. Introduction of Provider Mechanism

v2 introduces a brand new Provider mechanism, replacing the Vampire concept from v1:

### 4.1 Generic Provider Interface

```go
type Provider[T any] interface {
    Goner
    Provide(tagConf string) (T, error)
}
```

### 4.2 NamedProvider Interface

```go
type NamedProvider interface {
    NamedGoner
    Provide(tagConf string, t reflect.Type) (any, error)
}
```

### 4.3 NoneParamProvider Interface

```go
type NoneParamProvider[T any] interface {
    Goner
    Provide() T
}
```

The Provider mechanism allows components to dynamically create and provide other components, greatly enhancing the framework's flexibility and extensibility.

## 5. Multiple Instance Support

v2 enhances support for multiple instances, allowing creation of multiple Gone framework instances within the same application:

```go
// Create multiple Gone framework instances
app1 := gone.NewApp()
app2 := gone.NewApp()

// Each instance can independently load components and run
app1.Load(&Component1{})
app2.Load(&Component2{})

app1.Run()
app2.Run()
```

## 6. Dynamic Component Retrieval

v2 provides more flexible ways to dynamically retrieve components:

```go
type GonerKeeper interface {
    GetGonerByName(name string) any
    GetGonerByType(t reflect.Type) any
}
```

## 7. Function Parameter Injection

v2 enhances function parameter injection functionality:

```go
type FuncInjector interface {
    InjectWrapFunc(fn interface{}, args []interface{}, kwargs map[string]interface{}) (func() error, error)
}
```

## 8. Repository Structure Optimization

v2 has made important repository structure adjustments, separating `github.com/gone-io/gone/goner` into an independent repository, while the `github.com/gone-io/gone` repository focuses on managing Gone's core dependency injection code:

```
// v1 version
github.com/gone-io/gone        // Contains all Gone framework code
github.com/gone-io/gone/goner  // As a subdirectory of the main repository

// v2 version
github.com/gone-io/gone       // Contains only dependency injection core code
github.com/gone-io/goner      // Independent repository, manages Goner-related code
```

This repository structure optimization brings the following benefits:

1. **Clearer module boundaries**: By making Goner-related code independent, the framework's module boundaries become clearer, with each repository having specific responsibilities and functional scope.

2. **More flexible version management**: Independent repositories can have independent release cycles, allowing the Goner module to iterate and update according to its own needs, without synchronizing with the main framework.

3. **Better code reusability**: The independent Goner repository can be more conveniently referenced and reused by other projects without importing the entire Gone framework.

4. **More focused maintenance responsibilities**: Teams can divide maintenance of different repositories according to expertise, improving development efficiency and code quality.

5. **Reduced dependency complexity**: Users can selectively import the modules they need based on actual requirements, reducing unnecessary dependencies.

This repository structure adjustment reflects Gone framework's continuous optimization of architectural design, making the framework more modular and maintainable.

## 9. Migration Guide

When migrating from v1 to v2, pay attention to the following points:

1. **Update import paths**: Use `github.com/gone-io/gone/v2` instead of `github.com/gone-io/gone`

2. **Adjust component definitions**: Ensure all components embed `gone.Flag`

3. **Use new loading methods**: Adopt the component loading approaches provided in v2

4. **Adapt to the new Provider mechanism**: If using custom Providers, adjust to the v2 Provider interfaces

5. **Check lifecycle methods**: Ensure lifecycle methods conform to v2 specifications

## 10. Summary

Gone v2 makes the framework more usable, flexible, and powerful through improvements in several areas:

1. **Concept simplification**: Removing religious terminology, using more intuitive technical terms
2. **Interface redesign**: Reducing exposure of internal methods, making interfaces clearer
3. **Component loading mechanism improvement**: Providing more consistent and flexible component loading approaches
4. **Provider mechanism introduction**: Replacing the Vampire concept, offering more powerful component creation and provision capabilities
5. **Lifecycle management optimization**: Making component lifecycles clearer and more predictable
6. **Multiple instance support**: Allowing creation of multiple Gone framework instances within the same application
7. **Dynamic component retrieval**: Providing more flexible ways to dynamically retrieve components
8. **Function parameter injection**: Enhancing function parameter injection functionality
9. **Repository structure optimization**: Making Goner an independent repository, making the framework more modular and maintainable

These improvements make the Gone framework more suitable for building complex applications, especially microservice architecture applications.