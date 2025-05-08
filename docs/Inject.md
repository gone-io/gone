<p>
   English&nbsp ｜&nbsp <a href="./inject_CN.md">中文</a>
</p>

# Dependency Injection Introduction

- [Dependency Injection Introduction](#dependency-injection-introduction)
  - [Dependency Injection Categories](#dependency-injection-categories)
    - [By Injected Object Type](#by-injected-object-type)
    - [By Source of Injected Object](#by-source-of-injected-object)
    - [By Type of Injected Object](#by-type-of-injected-object)
    - [By Injection Method](#by-injection-method)
  - [What is a Goner?](#what-is-a-goner)
    - [`gone` Tag Format](#gone-tag-format)
    - [Code Example](#code-example)
  - [How to Register Goner to Gone Framework?](#how-to-register-goner-to-gone-framework)
    - [Method 1: Single Registration](#method-1-single-registration)
    - [Method 2: Batch Registration](#method-2-batch-registration)
  - [Dependency Injection Execution Timing](#dependency-injection-execution-timing)
  - [Manual Dependency Injection](#manual-dependency-injection)
    - [Method 1: Struct Injection (StructInjector)](#method-1-struct-injection-structinjector)
    - [Method 2: Function Parameter Injection (FuncInjector)](#method-2-function-parameter-injection-funcinjector)
  - [Summary](#summary)


When using the Gone framework, you might wonder how it performs dependency injection. This article will take you deep into Gone's injection mechanism and help you master its usage through examples.

---

## Dependency Injection Categories

### By Injected Object Type
1. Injected object is a struct field

As shown in the following code, we want the framework to automatically inject the implementation class of `Info` interface.
```go
type Info interface {
    GetInfo() string
}

type Dep struct {
    gone.Flag
    Name Info `gone:"*"` // Field to be injected
}
```
2. Injected object is a function parameter

As shown in the following code, we want the framework to automatically fill in the required parameters when calling a function.
```go
func hello(
    info Info, // Object to be injected
) {
    println(info.GetInfo())
}

func main() {
    gone.Run(hello)
}
```

### By Source of Injected Object
1. From objects registered to the framework:
```go
type Dep struct {
    gone.Flag
}

type UseDep struct {
    gone.Flag
    dep *Dep `gone:"*"` // Field to be injected, from framework-registered object
}
```
2. From system configuration parameters like environment variables, config files, etc.
```go
type UseDep struct {
    gone.Flag
    name string `gone:"config:name"` // Field to be injected, from config
}
```
3. From third-party components
```go
type UseDep struct {
    gone.Flag
    redis *redis.Client `gone:"*"` // Field to be injected, from third-party
}
```

### By Type of Injected Object
1. Pointer type
2. Interface type
3. Struct type
4. Basic type

### By Injection Method
1. Anonymous injection
```go
type UseDep struct {
    gone.Flag
    dep *Dep `gone:"*"` // Field to be injected, matched by type
}
```
2. Named injection
```go
type UseDep struct {
    gone.Flag
    dep *Dep `gone:"dep-name"` // Field to be injected, requires specific name
}
```

---

## What is a Goner?

In the Gone framework, injected objects are called **Goners**. To make an object a Goner, it must meet two conditions:

1. **The object must embed `gone.Flag`** - This Flag lets Gone recognize it as a Goner.
2. **Fields needing injection must be tagged with `gone`** - Only fields with `gone` tag will be injected.

### `gone` Tag Format

```go
gone:"${name},${extend}"
```

- **`${name}`**: Goner name. If `*` or omitted, indicates automatic injection by type.
- **`${extend}`** (optional): Extended options passed to Provider's `Provide` method, explained in [Gone V2 Provider Mechanism](./provider.md).

### Code Example

```go
type Dep struct {
    gone.Flag
}

type Dep2 struct {
    gone.Flag
}

type UseDep struct {
    gone.Flag
    dep  *Dep  `gone:"*"`   // Auto-injected by type
    Dep2 *Dep2 `gone:"dep2"` // Injected by name "dep2"
}
```

> **Fields can be private!** This follows the **Open-Closed Principle**, making it safer and more encapsulated.

---

## How to Register Goner to Gone Framework?

Registering Goners is simple, typically done in two ways:

### Method 1: Single Registration

```go
gone.Load(&UseDep{})
```

### Method 2: Batch Registration

If you need to register multiple Goners at once:

```go
gone.Loads(func(l gone.Loader) error {
    _ = l.Load(&UseDep{})
    _ = l.Load(&Dep{})
    _ = l.Load(&Dep2{}, gone.Name("dep2"))
    return nil
})
```

You can pass additional parameters during registration, like `gone.Name()` to specify Goner name.

---

## Dependency Injection Execution Timing

After framework startup, all registered objects automatically undergo dependency injection. If any `gone` tag cannot find its dependency, the framework will report an error to ensure dependency integrity.

---

## Manual Dependency Injection

Although Gone can inject automatically, sometimes we want **manual control**. The framework provides two manual injection methods:

1. **`gone.StructInjector`** - For struct field injection.
2. **`gone.FuncInjector`** - For function parameter injection.

### Method 1: Struct Injection (StructInjector)

Suppose we have a `Business` struct that depends on `Dep`, but `Dep` doesn't exist initially and needs runtime injection.

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

### Method 2: Function Parameter Injection (FuncInjector)

If you have a function `needInjectedFunc` that depends on `Dep`, you can use `gone.FuncInjector` to auto-inject its parameters.

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

## Summary

- **Gone framework uses `gone.Flag` and `gone` tags for auto dependency injection.**
- **Goners can be injected by type (`*`) or by name (`gone:"dep2"`).**
- **Objects must be registered to framework before participating in injection.**
- **If auto injection isn't enough, use `gone.StructInjector` and `gone.FuncInjector` for manual injection.**

With these fundamentals, you can efficiently use Gone's dependency injection mechanism in your projects! 🎉