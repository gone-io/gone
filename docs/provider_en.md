# Gone V2 Provider Mechanism Introduction

- [Gone V2 Provider Mechanism Introduction](#gone-v2-provider-mechanism-introduction)
  - [1. Gone's Dependency Injection Process](#1-gones-dependency-injection-process)
  - [2. Different Providers](#2-different-providers)
    - [2.1 Injection by Type](#21-injection-by-type)
    - [2.2 Injection by Name](#22-injection-by-name)
    - [2.3 Multi-type Provider Based on Name](#23-multi-type-provider-based-on-name)
  - [3. "Asterisk" Provider: `*`](#3-asterisk-provider-)
  - [Summary](#summary)


The Gone V2 version is completely based on the Provider mechanism to implement dependency injection. Its core idea is to provide dependencies for objects through Providers, without having to wrap all third-party objects as `Goner`. The following will explain in detail from aspects such as Gone's dependency injection process, Provider classification, and usage examples.

---

## 1. Gone's Dependency Injection Process

The dependency injection process of the Gone framework is mainly divided into the following three steps:

1. **Mark objects that need injection**  
   Objects need to embed `gone.Flag` and use the `gone` tag for fields that need injection. For example:
   ```go
   type UseConf struct {
   	gone.Flag
   	Name string `gone:"config,name"`
   	Int  string `gone:"config",int`
   	Dep  *Dep   `gone:"*,extend"`
   }
   ```
    - `gone` tag format: `gone:"${name},${extend}"`
        - `${name}`: Specifies the name of the Provider or the name of the target `Goner` (when it's `*` or omitted, it indicates injection by type).
        - `${extend}`: Extension parameter, which will be passed to the Provider's `Provide` method (optional).

2. **Register objects**  
   Register objects that need to be injected into the Gone framework through `Load(goner Goner, options ...Option)`.

3. **Automatic injection and lifecycle management**  
   During the framework startup process, automatic dependency injection is performed for all registered objects, while lifecycle methods such as `BeforeInit`, `Init`, `Start`, `Stop` are called (if the object implements the corresponding interface).

> **Note:** The Gone framework only allows registration of objects that implement the `Goner` interface, and typically implements the `Goner` interface by embedding `gone.Flag`. This distinguishes which objects need injection and which don't. But this also brings up a question: how to inject third-party objects into the framework? For this purpose, the Provider mechanism was introduced.

---

## 2. Different Providers

Gone mainly supports two injection methods: **injection by type** and **injection by name**.

### 2.1 Injection by Type

When no specific name is specified for a field (using `*` or omitting the name in the tag), the framework will look for a suitable Provider based on the field's type. Two Provider interfaces are defined here:

- **Provider with parameter support**
  ```go
  type Provider[T any] interface {
      Goner
      Provide(tagConf string) (T, error)
  }
  ```

- **Provider without parameters**
  ```go
  type NoneParamProvider[T any] interface {
      Goner
      Provide() (T, error)
  }
  ```

**Example code:**
```go
package main

import "github.com/gone-io/gone/v2"

type ThirdBusiness1 struct {
    Name string
}

type ThirdBusiness2 struct {
    Name string
}

// Provider implementation with parameters
type ThirdBusiness1Provider struct {
    gone.Flag
    gone.Logger `gone:"*"`
}

func (p *ThirdBusiness1Provider) Provide(tagConf string) (*ThirdBusiness1, error) {
    p.Infof("tagConf->%s", tagConf)
    return &ThirdBusiness1{Name: "ThirdBusiness1"}, nil
}

// Provider implementation without parameters
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

Execution result:
```log
2025/03/11 10:03:22 tagConf->AGI
2025/03/11 10:03:22 user.thirdBusiness1.name->ThirdBusiness1
2025/03/11 10:03:22 user.thirdBusiness2.name->ThirdBusiness2
```

### 2.2 Injection by Name

When there are multiple Providers of the same type, they can be distinguished by name. The `gone` tag on the field specifies the Provider name, requiring that the object type returned by the Provider is compatible with the field type.

**Example code:**
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

There are two ways to specify a name when registering a Provider:
1. **Implement the `GonerName()` method** — Returns the name of the Provider.
2. **Use the `gone.Name("...")` option when `Load`ing** — Explicitly specify the Provider name.

> **Tip:** When providing multiple Providers of the same type, you need to use the `gone.OnlyForName()` option, otherwise the framework will report an error; and only one Provider with the same name can exist, otherwise an error will also be reported.

### 2.3 Multi-type Provider Based on Name

In some scenarios (such as configuration injection), you may want to provide objects of multiple types through one Provider. In this case, you can define the `NamedProvider` interface:
```go
type NamedProvider interface {
    NamedGoner
    Provide(tagConf string, t reflect.Type) (any, error)
}
```
Where:
- The `tagConf` parameter is the extension configuration part after the first comma in the `gone` tag.
- The `t` parameter represents the type of field that needs to be injected.

This design allows a Provider to return corresponding instances based on the field type, implementing multi-type injection.

---

## 3. "Asterisk" Provider: `*`

When the name in the `gone` tag is `*` or omitted, it means injection by type. This is actually a predefined `NamedProvider` in the framework, whose name is `*`.  
Its working logic is:
- Automatically find and call the appropriate Provider to provide values based on the type of field that needs to be injected.

---

## Summary

Gone V2 implements flexible dependency injection through the Provider mechanism, supporting:

- **Injection by type**: Directly look for the appropriate Provider based on the field type.
- **Injection by name**: Distinguish by name when there are multiple Providers of the same type.
- **Multi-type Provider**: One Provider can return different objects according to the field type, suitable for scenarios such as configuration injection.
- **"Asterisk" Provider**: Simplifies the processing flow of injection by type.

This design not only reduces the coupling between third-party objects and the framework but also greatly improves the flexibility and scalability of dependency injection.

---