# Introduction to Dependency Injection

- [Introduction to Dependency Injection](#introduction-to-dependency-injection)
  - [What is a Goner?](#what-is-a-goner)
    - [Format of the `gone` tag](#format-of-the-gone-tag)
    - [Code Example](#code-example)
  - [How to Register Goners in the Gone Framework?](#how-to-register-goners-in-the-gone-framework)
    - [Method 1: Individual Registration](#method-1-individual-registration)
    - [Method 2: Batch Registration](#method-2-batch-registration)
  - [When Does Dependency Injection Happen?](#when-does-dependency-injection-happen)
  - [Manual Dependency Injection](#manual-dependency-injection)
    - [Method 1: Struct Injection (StructInjector)](#method-1-struct-injection-structinjector)
    - [Method 2: Function Parameter Injection (FuncInjector)](#method-2-function-parameter-injection-funcinjector)
  - [Summary](#summary)


When using the Gone framework, you might wonder how it handles dependency injection. This article will guide you through Gone's injection mechanism and help you master its usage through examples.

---

## What is a Goner?

In the Gone framework, objects that can be injected are called **Goners**. To make an object a Goner, it must meet two conditions:

1. **The object must embed `gone.Flag`** — This Flag helps Gone recognize it as a Goner.
2. **Fields requiring injection must be tagged with `gone`** — Only fields tagged with `gone` will be injected.

### Format of the `gone` tag

```go
gone:"${name},${extend}"
```

- **`${name}`**: Represents the Goner's name. If it's `*` or omitted, it means automatic injection by type.
- **`${extend}`** (optional): Extension options that will be passed to the Provider's `Provide` method, which is explained in more detail in ["Gone V2 Provider Mechanism Introduction"](./provider_en.md).

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
    dep  *Dep  `gone:"*"`    // Automatically injected by type
    Dep2 *Dep2 `gone:"dep2"` // Injected by name "dep2"
}
```

:::tip
**Fields can be private!** This follows the **Open/Closed Principle**, making it safer and more aligned with encapsulation design.
:::

---

## How to Register Goners in the Gone Framework?

Registering Goners is quite simple and typically done in two ways:

### Method 1: Individual Registration

```go
gone.Load(&UseDep{})
```

### Method 2: Batch Registration

If you need to register multiple Goners at once, you can do it like this:

```go
gone.Loads(func(l gone.Loader) error {
    _ = l.Load(&UseDep{})
    _ = l.Load(&Dep{})
    _ = l.Load(&Dep2{}, gone.Name("dep2"))
    return nil
})
```

When registering, you can also pass multiple extension parameters, such as `gone.Name()` to specify the Goner's name.

---

## When Does Dependency Injection Happen?

After the framework starts, all registered objects will automatically undergo dependency injection. If a field's `gone` tag can't find its corresponding dependency, the framework will immediately report an error, ensuring dependency integrity.

---

## Manual Dependency Injection

Although Gone can inject dependencies automatically, sometimes we want to **manually control** the injection process. The framework provides two methods for manual injection:

1. **`gone.StructInjector`** — Used for injecting into struct fields.
2. **`gone.FuncInjector`** — Used for injecting function parameters.

### Method 1: Struct Injection (StructInjector)

Suppose we have a `Business` struct that depends on `Dep`, but `Dep` doesn't exist initially and needs to be injected at runtime.

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

If you have a function `needInjectedFunc` that requires `Dep`, you can use `gone.FuncInjector` to automatically inject its parameters.

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

- **The Gone framework uses `gone.Flag` and the `gone` tag for automatic dependency injection.**
- **Goners can be injected by type (`*`) or by name (`gone:"dep2"`).**
- **Objects must be registered with the framework first to participate in dependency injection.**
- **If automatic injection isn't sufficient, you can use `gone.StructInjector` and `gone.FuncInjector` for manual injection.**

After mastering these basic concepts, you'll be able to efficiently use Gone framework's dependency injection mechanism in your projects! 🎉