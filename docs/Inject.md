<p>
   English&nbsp ｜&nbsp <a href="./inject_CN.md">中文</a>
</p>

- [In-depth Understanding of Gone Framework's Dependency Injection Mechanism](#in-depth-understanding-of-gone-frameworks-dependency-injection-mechanism)
  - [What is Dependency Injection?](#what-is-dependency-injection)
    - [Core Concepts](#core-concepts)
    - [`gone` Tag Format](#gone-tag-format)
    - [Function Parameter Injection](#function-parameter-injection)
  - [Multiple Classifications of Dependency Injection](#multiple-classifications-of-dependency-injection)
    - [Classification by Injection Recipient Type](#classification-by-injection-recipient-type)
    - [Classification by Injection Matching Method](#classification-by-injection-matching-method)
    - [Classification by Injected Object Source](#classification-by-injected-object-source)
  - [Gone Framework Startup Process](#gone-framework-startup-process)
    - [`gone.NewApp` Method](#gonenewapp-method)
    - [`gone.Application`'s `Run` Method](#goneapplications-run-method)
    - [`gone.Application`'s `Serve` Method](#goneapplications-serve-method)
    - [`gone.Default` Default Instance](#gonedefault-default-instance)
  - [Loading Objects into Gone Framework](#loading-objects-into-gone-framework)
    - [`gone.Loader` and `gone.LoadFunc`](#goneloader-and-goneloadfunc)
    - [Comprehensive Example of Multiple Loading Methods](#comprehensive-example-of-multiple-loading-methods)
  - [Manual Control of Dependency Injection](#manual-control-of-dependency-injection)
    - [Struct Injection (StructInjector)](#struct-injection-structinjector)
    - [Function Parameter Injection (FuncInjector)](#function-parameter-injection-funcinjector)
  - [Conclusion](#conclusion)


# In-depth Understanding of Gone Framework's Dependency Injection Mechanism

Dependency injection is a core concept in modern software architecture that enables us to build loosely coupled, easily testable, and highly maintainable applications. The Gone framework provides an elegant and powerful dependency injection mechanism that allows us to easily manage complex component dependencies. This article will guide you through Gone framework's injection mechanism, helping you master this powerful design pattern through clear concept explanations and practical examples.

## What is Dependency Injection?

Dependency injection is a design pattern that allows us to decouple object dependencies from the code, letting the framework handle the assembly of these dependencies. This approach offers several important advantages: it reduces coupling between components, improves code testability, and simplifies object creation and management.

Let's understand dependency injection in the Gone framework through a simple example:

```go
type Dep struct {
    gone.Flag
}

type Service struct {
    gone.Flag
    dep *Dep `gone:"*"`
}
```

In this example, `Service` depends on `*Dep`. In traditional programming, we would need to manually initialize `*Dep` before using `Service`, which leads to tight coupling and makes testing difficult. However, in the Gone framework, dependency injection means the framework automatically identifies and completes the initialization of all fields marked with `gone` during startup, greatly simplifying the code and improving maintainability.

### Core Concepts

To better understand Gone framework's dependency injection mechanism, we need to clarify several key concepts:

- **Injected Object**: `Dep` is injected into `Service`, making `Dep` the injected object that provides functionality to other components
- **Injection Recipient**: `Service` receives the dependency injection of `Dep`, making `Service` the injection recipient that uses the injected object's functionality
- **Injected Field**: `Service.dep` is the injected field, and `*Dep` is the injection type, which the framework uses to find suitable objects for injection
- **Injection Tag**: `gone:"*"` is the injection tag, which not only identifies fields that need injection but can also provide additional control information for the injection process

### `gone` Tag Format

The Gone framework uses struct tags to control dependency injection behavior. The `gone` tag follows a specific format rule: `gone:"${pattern}[,${extend}]"`

- **`${pattern}`**: Injection matching pattern, which can be a specific name or a pattern string containing wildcards `*` or `?` for flexible target object matching
- **`${extend}`** (optional): Extension options passed to the Provider's `Provide` method for additional injection control. More details can be found in ["Gone V2 Provider Mechanism Introduction"](./provider_CN.md)

This design allows developers to precisely control dependency injection behavior, meeting the needs of various complex scenarios.

### Function Parameter Injection

Gone framework's dependency injection capability extends beyond struct fields to function parameters. The framework provides a mechanism for wrapping functions (currying) that allows some or all function parameters to be automatically filled with framework-managed objects. This is particularly useful when handling callbacks, middleware, or event handlers.

Here's a test case demonstrating function parameter injection:

```go
package use_case

import (
  "github.com/gone-io/gone/v2"
  "testing"
)

func TestFuncInject(t *testing.T) {
  type Dep struct {
    gone.Flag
    ID int
  }

  fnExecuted := false
  fn := func(dep *Dep) {
    if dep.ID != 1 {
      t.Fatal("func inject failed")
    }
    fnExecuted = true
  }

  gone.
    NewApp().
    Load(&Dep{ID: 1}).
          Run(func(injector gone.FuncInjector) {
            wrapFunc, err := injector.InjectWrapFunc(fn, nil, nil)
            if err != nil {
              t.Fatal(err)
            }
            _ = wrapFunc()
          })

  if !fnExecuted {
    t.Fatal("func inject failed")
  }
}
```

For function parameter injection, we have similar concept mappings:

- `Dep` is injected into function `fn`, making `Dep` the injected object
- `fn` receives the dependency injection of `Dep`, making `fn` the injection recipient
- `fn`'s parameter `dep` is the injected field, and `*Dep` is the injection type

This function parameter injection mechanism enables the Gone framework to provide flexible support for various programming patterns, especially in handling callbacks and event-driven scenarios.

## Multiple Classifications of Dependency Injection

Gone framework's dependency injection system is highly flexible and diverse, with multiple dimensions of classification to accommodate different usage scenarios and requirements.

### Classification by Injection Recipient Type

1. **Struct Injection**

   In the Gone framework, structs receiving injection need to anonymously embed `gone.Flag`. This embedding makes the struct pointer implement the `gone.Goner` interface, which we call a **Goner**:

   ```go
   type Service struct {
       gone.Flag
       dep *Dep `gone:"*"` // Field requiring injection
   }
   ```

   This design allows the framework to identify and manage structs, providing them with dependency injection capabilities.

2. **Function Parameter Injection**

   As mentioned earlier, Gone framework supports injecting function parameters, allowing functions to automatically obtain required dependency objects. This mechanism is particularly useful in scenarios like HTTP request handlers and event listeners.

### Classification by Injection Matching Method

Gone framework provides multiple matching methods to make the dependency injection process more flexible and precise:

1. **Name Matching**

   Through the `${pattern}` in the `gone` tag, you can specify the name of the injected object for exact matching:

   ```go
   type UseDep struct {
       gone.Flag
       dep *Dep `gone:"dep-name"` // Field requiring injection, needs an object with specific name
   }
   ```

   This method is suitable for situations where you need to distinguish between multiple instances of the same type, such as clients connecting to different databases.

2. **Type Matching**

   When the `gone` tag value omits `${pattern}` or sets it to `*`, it indicates matching one value by type:

   ```go
   type UseDep struct {
       gone.Flag
       dep *Dep `gone:"*"` // Field requiring injection, matches one value by type
       dep1 *Dep `gone:""` // `gone:""` is equivalent to `gone:"*"`
       dep2 *Dep `gone:"*,extend-value"` // `${extend}` is not empty
       dep3 *Dep `gone:",extend-value"` // `gone:",extend-value"` is equivalent to `gone:"*,extend-value"`
   }
   ```

   This is the most commonly used matching method, suitable for most simple dependency relationships.

3. **Type and Wildcard Matching**

   The `${pattern}` in the `gone` tag can also be a pattern string containing wildcards for more flexible matching rules:

   ```go
   type UseDep struct {
       gone.Flag
       dep *Dep `gone:"dep-*-?"` 
   }
   ```

   Here, `*` matches any number of characters, and `?` matches a single character. This method provides greater flexibility for component naming and selection.

   > It's worth noting that both **Name Matching** and **Type Matching** can actually be seen as special cases of **Type and Wildcard Matching**. Gone framework provides this unified and flexible matching mechanism to make the dependency injection process more powerful and adaptable.

### Classification by Injected Object Source

Gone framework supports obtaining injected objects from multiple sources, further enhancing the framework's flexibility:

1. **From Framework-Registered Objects**

   The most common case is injecting objects registered within the framework:

   ```go
   type Dep struct {
       gone.Flag
   }
   
   type UseDep struct {
       gone.Flag
       dep *Dep `gone:"*"` // Field requiring injection, sourced from framework-registered object
   }
   ```

   This method is suitable for most application scenarios, allowing the framework to manage component lifecycles and dependencies.

2. **From System Configuration Parameters**

   Gone framework supports injecting values from external sources like environment variables, configuration files, and configuration centers, making configuration management more flexible:

   ```go
   type UseDep struct {
       gone.Flag
       name string `gone:"config:name"` // Field requiring injection, sourced from configuration parameters
   }
   ```

   This method allows applications to adapt to different runtime environments without code modifications.

3. **From Third-Party Components**

   The framework also supports injecting third-party components, enabling seamless integration with external systems:

   ```go
   type UseDep struct {
       gone.Flag
       redis *redis.Client `gone:"*"` // Field requiring injection, sourced from third-party component
   }
   ```

   This capability allows the Gone framework to work with various external libraries and services, expanding the application's functionality.

## Gone Framework Startup Process

Gone framework provides multiple ways to start and manage applications, allowing developers to choose flexibly based on different needs.

### `gone.NewApp` Method

Through `gone.NewApp`, you can create a `gone.Application` instance, which is the core container of the application:

```go
app := gone.NewApp()
```

After creating the instance, you can use the `Application::Load` and `Application::Loads` methods to load Goner objects into the framework, preparing for dependency injection.

### `gone.Application`'s `Run` Method

After loading all Goner objects, you can start the framework using the `Application::Run` method:

```go
app.Run(func(service *MyService) {
    service.DoSomething()
})
```

The `Run` method supports passing multiple functions as parameters, which will be executed in sequence and support function parameter injection. This design is particularly suitable for executing a series of initialization tasks or starting multiple parallel services.

### `gone.Application`'s `Serve` Method

For services that need to run for extended periods, Gone framework provides the `Serve` method:

```go
app.Serve()
```

`Application::Serve` is similar to the `Application::Run` method, but `Serve` blocks the current thread until the service receives a stop signal or `Application::End` is called to manually stop the service. This method is particularly suitable for background service programs and web applications.

Note that the `Serve` method doesn't support passing parameters because it's primarily used for starting fully initialized long-running services.

### `gone.Default` Default Instance

To simplify usage, Gone framework provides a default `Application` instance, which can be operated directly through the following global methods:

- `gone.Run` - Run the application
- `gone.Serve` - Start long-running services
- `gone.End` - Stop services
- `gone.Load` - Load a single Goner object
- `gone.Loads` - Load multiple Goner objects

This design makes writing simple applications more concise, without the need to explicitly create application instances.

## Loading Objects into Gone Framework

Gone framework provides multiple flexible ways to load objects, allowing developers to choose the most suitable method based on project organization structure and dependency relationships.

### `gone.Loader` and `gone.LoadFunc`

`gone.Loader` is a core interface provided by Gone framework for loading objects, defining how to load objects into the framework:

```go
type Loader interface {
    Load(goner Goner, options ...Option) error
    MustLoad(goner Goner, options ...Option) Loader // Load goner, panic if loading fails, supports chain calls
    MustLoadX(x any) Loader // Load x, x can be Goner or LoadFunc
    Loaded(LoaderKey) bool
}
```

And `gone.LoadFunc` is a function type that defines the loading function:

```go
type LoadFunc = func(Loader) error
```

This design allows us to encapsulate business logic as components, with each component potentially containing multiple injectable objects. By encapsulating loading logic into `LoadFunc` functions, related objects can be conveniently loaded into the framework together:

```go
package componentA
import "github.com/gone-io/gone/v2"

type A struct {
    gone.Flag
}

func ALoad(loader gone.Loader) error {
    // Load component B dependencies
    loader.MustLoadX(componentB.BLoad)
    
    // Load component A related objects
    loader.MustLoad(&A{})

    return nil
}
```

More powerfully, this design supports dependencies between components: if component A depends on component B, you can first load component B in component A's `LoadFunc` function, ensuring dependencies are ready before being depended upon.

### Comprehensive Example of Multiple Loading Methods

Below is a comprehensive example showing Gone framework's multiple object loading methods:

```go
package main

import "github.com/gone-io/gone/v2"
import "fmt"

type A struct {
    gone.Flag
    ID string
}

func main() {
    gone.
        NewApp(
            // 1. Load objects using `gone.NewApp`, supports multiple `LoadFunc` functions as parameters
        ).
        // 2. Load objects using chain calls of `Application` instance methods
        Load(&A{ID: "a"}, gone.Name("instance-a")).
        Load(&A{ID: "b"}, gone.Name("instance-b")).
        // 3. Load objects using `Application` instance's `Loads` method through `LoadFunc` method
        Loads(
            func(loader gone.Loader) error {
                // 4. Load objects using chain calls of `gone.Loader`
                loader.
                    MustLoad(&A{ID: "c"}, gone.Name("instance-c")).
                    MustLoad(&A{ID: "d"}, gone.Name("instance-d")).
                    // 5. Load objects by calling other `LoadFunc` methods through `gone.Loader::MustLoadX`
                    MustLoadX(func(loader gone.Loader) error {
                        return loader.Load(&A{ID: "f"}, gone.Name("instance-f"))    
                    })
                
                return loader.Load(&A{ID: "e"}, gone.Name("instance-e"))
            },
            // Supports multiple `LoadFunc` functions as parameters
        ).
        Run(func(a []*A) {
            fmt.Printf("%#v", a)
        })
        // Can also use Serve() to start long-running services
}
```

This example demonstrates multiple ways of loading objects, from simple direct loading to complex chain and nested loading, satisfying different organizational structure and dependency relationship needs.

## Manual Control of Dependency Injection

While Gone framework's automatic dependency injection mechanism can handle most scenarios, sometimes we may need to manually control the injection process in specific situations. For this, the framework provides two specialized interfaces:

### Struct Injection (StructInjector)

The `gone.StructInjector` interface is used for manually injecting struct fields. This is particularly useful when creating objects dynamically at runtime:

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
    // Manually complete struct field injection
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

In this example, the `Business` class obtains a `StructInjector` and uses it to manually inject fields of the `User` struct. This method allows for dynamic object creation and injection at runtime, particularly suitable for handling user input or configuration-driven scenarios.

### Function Parameter Injection (FuncInjector)

The `gone.FuncInjector` interface is used for manually injecting function parameters. This is particularly useful when handling callbacks, middleware, or event handlers:

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
    
    // Manually complete function parameter injection
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

In this example, the `Business` class obtains a `FuncInjector` and uses it to wrap the function `needInjectedFunc`, automatically injecting function parameters. This method makes handling callbacks and events more concise and flexible.

## Conclusion

Gone framework provides a powerful and flexible dependency injection mechanism that enables Go language developers to build loosely coupled, testable, and maintainable applications. Through struct field injection and function parameter injection, the framework meets the needs of various complex scenarios. Additionally, the framework's multiple loading methods and manual control mechanisms further enhance developers' flexibility and control.

Dependency injection is an important concept in modern software development that changes how we organize and manage code. Through Gone framework's dependency injection mechanism, we can more easily build large, complex applications without worrying about tight coupling between components or testing difficulties.

As you deepen your understanding and practice with the Gone framework, you'll be able to fully leverage the advantages of dependency injection to build more robust and maintainable Go applications.
