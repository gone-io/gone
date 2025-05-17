<p>
    <a href="inject.md">English</a>&nbsp ｜&nbsp 中文
</p>

- [深入理解Gone框架的依赖注入机制](#深入理解gone框架的依赖注入机制)
  - [什么是依赖注入？](#什么是依赖注入)
    - [核心概念](#核心概念)
    - [`gone`标签的格式](#gone标签的格式)
    - [函数参数注入](#函数参数注入)
  - [依赖注入的多种分类](#依赖注入的多种分类)
    - [按接受注入对象的类型分类](#按接受注入对象的类型分类)
    - [按注入时的匹配方式分类](#按注入时的匹配方式分类)
    - [按被注入对象的来源分类](#按被注入对象的来源分类)
  - [Gone框架的启动流程](#gone框架的启动流程)
    - [`gone.NewApp`方法](#gonenewapp方法)
    - [`gone.Application`的`Run`方法](#goneapplication的run方法)
    - [`gone.Application`的`Serve`方法](#goneapplication的serve方法)
    - [`gone.Default`默认实例](#gonedefault默认实例)
  - [将对象加载到Gone框架](#将对象加载到gone框架)
    - [`gone.Loader`和`gone.LoadFunc`](#goneloader和goneloadfunc)
    - [多种加载方式的综合示例](#多种加载方式的综合示例)
  - [手动控制依赖注入](#手动控制依赖注入)
    - [结构体注入（StructInjector）](#结构体注入structinjector)
    - [函数参数注入（FuncInjector）](#函数参数注入funcinjector)
  - [总结](#总结)


# 深入理解Gone框架的依赖注入机制

依赖注入是现代软件架构中的一个核心概念，它能够让我们构建松耦合、易测试且高度可维护的应用程序。Gone框架提供了一套优雅而强大的依赖注入机制，让我们能够轻松管理复杂的组件依赖关系。本文将带您深入了解Gone框架的注入机制，通过清晰的概念解释和实用的示例，帮助您掌握这一强大的设计模式。

## 什么是依赖注入？

依赖注入是一种设计模式，它允许我们将对象的依赖关系从代码中解耦出来，由框架负责组装这些依赖。这样做有几个重要优势：降低了组件间的耦合度，提高了代码的可测试性，并简化了对象的创建和管理过程。

让我们通过一个简单的例子来理解Gone框架中的依赖注入：

```go
type Dep struct {
    gone.Flag
}

type Service struct {
    gone.Flag
    dep *Dep `gone:"*"`
}
```

在这个例子中，`Service`依赖于`*Dep`。在传统编程中，我们需要在使用`Service`之前手动初始化`*Dep`，这会导致代码紧密耦合且难以测试。而在Gone框架中，依赖注入意味着框架会在启动时自动识别并完成所有带有`gone`标记的字段的初始化过程，大大简化了代码并提高了可维护性。

### 核心概念

为了更好地理解Gone框架的依赖注入机制，我们需要明确几个关键概念：

- **被注入对象**：`Dep`被注入到`Service`中，因此`Dep`是被注入对象，它为其他组件提供功能
- **接受注入的对象**：`Service`接受了`Dep`的依赖注入，因此`Service`是接受注入的对象，它使用被注入对象的功能
- **接受注入的字段**：`Service.dep`是接受注入的字段，而`*Dep`是注入类型，框架会根据这个类型来查找合适的对象进行注入
- **注入标记**：`gone:"*"`是注入标记，它不仅标识需要注入的字段，还可以为注入过程提供额外的控制信息

### `gone`标签的格式

Gone框架使用结构体标签（struct tag）来控制依赖注入的行为。`gone`标签遵循特定的格式规则：`gone:"${pattern}[,${extend}]"`

- **`${pattern}`**：注入匹配模式，可以是一个具体的名称，也可以是包含通配符`*`或`?`的模式字符串，用于灵活匹配目标对象
- **`${extend}`**（可选）：扩展选项，会传递给Provider的`Provide`方法，为注入过程提供更多控制。在[《Gone V2 Provider 机制介绍》](./provider_CN.md)中有更详细的介绍

这种设计允许开发者精确控制依赖注入的行为，满足各种复杂场景的需求。

### 函数参数注入

Gone框架的依赖注入能力不仅限于结构体字段，还扩展到了函数参数。框架提供了一种将函数进行包装（柯里化）的机制，使函数的部分或全部参数能够自动填充为框架管理的对象。这在处理回调、中间件或事件处理器时特别有用。

下面是一个演示函数参数注入的测试用例：

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

对于函数参数注入，我们有类似的概念映射：

- `Dep`被注入到函数`fn`中，因此`Dep`是被注入对象
- `fn`接受了`Dep`的依赖注入，因此`fn`是接受注入的对象
- `fn`的参数`dep`是接受注入的字段，`*Dep`是注入类型

这种函数参数注入机制使Gone框架能够为各种编程模式提供灵活的支持，特别是在处理回调和事件驱动的场景中。

## 依赖注入的多种分类

Gone框架的依赖注入系统非常灵活多样，可以从多个维度进行分类，以适应不同的使用场景和需求。

### 按接受注入对象的类型分类

1. **结构体注入**

   在Gone框架中，接受注入的结构体需要匿名嵌入`gone.Flag`。这种嵌入使结构体的指针实际上实现了接口`gone.Goner`，我们称之为**Goner**：

   ```go
   type Service struct {
       gone.Flag
       dep *Dep `gone:"*"` // 需要注入的字段
   }
   ```

   这种设计使框架能够识别和管理结构体，并为其提供依赖注入的能力。

2. **函数参数注入**

   如前所述，Gone框架支持对函数参数进行注入，使函数能够自动获取所需的依赖对象。这种机制在许多场景下非常有用，比如HTTP请求处理器、事件监听器等。

### 按注入时的匹配方式分类

Gone框架提供了多种匹配方式，使依赖注入过程更加灵活和精确：

1. **名称匹配**

   通过`gone`标签的`${pattern}`可以指定被注入对象的名称，实现精确匹配：

   ```go
   type UseDep struct {
       gone.Flag
       dep *Dep `gone:"dep-name"` // 需要注入的字段，需要是特定名称的对象
   }
   ```

   这种方式适用于需要区分同一类型的多个实例的情况，比如连接到不同数据库的客户端。

2. **类型匹配**

   当`gone`标签的值省略`${pattern}`或设为`*`时，表示按类型匹配一个值：

   ```go
   type UseDep struct {
       gone.Flag
       dep *Dep `gone:"*"` // 需要注入的字段，按类型匹配一个值
       dep1 *Dep `gone:""` // `gone:""` 等价于 `gone:"*"`
       dep2 *Dep `gone:"*,extend-value"` // `${extend}` 不为空
       dep3 *Dep `gone:",extend-value"` // `gone:",extend-value"` 等价于 `gone:"*,extend-value"`
   }
   ```

   这是最常用的匹配方式，适用于大多数简单的依赖关系。

3. **类型和通配符匹配**

   `gone`标签的`${pattern}`也可以是包含通配符的模式字符串，支持更灵活的匹配规则：

   ```go
   type UseDep struct {
       gone.Flag
       dep *Dep `gone:"dep-*-?"` 
   }
   ```

   这里，`*`表示匹配任意多个字符，`?`表示匹配单个字符。这种方式为组件命名和选择提供了更大的灵活性。

   > 值得注意的是，**名称匹配**和**类型匹配**实际上都可以看作是**类型和通配符匹配**的特殊情况。Gone框架提供了这种统一而灵活的匹配机制，使依赖注入过程更加强大和适应性强。

### 按被注入对象的来源分类

Gone框架支持从多种来源获取被注入的对象，进一步增强了框架的灵活性：

1. **来源于注册到框架的对象**

   最常见的情况是注入来自框架内部注册的对象：

   ```go
   type Dep struct {
       gone.Flag
   }
   
   type UseDep struct {
       gone.Flag
       dep *Dep `gone:"*"` // 需要注入的字段，来源于注册到框架的对象
   }
   ```

   这种方式适用于大多数应用场景，使框架能够管理组件的生命周期和依赖关系。

2. **来源于系统配置参数**

   Gone框架支持从环境变量、配置文件、配置中心等外部源注入值，使配置管理更加灵活：

   ```go
   type UseDep struct {
       gone.Flag
       name string `gone:"config:name"` // 需要注入的字段，来源于配置参数
   }
   ```

   这种方式使应用程序能够适应不同的运行环境，无需修改代码即可改变行为。

3. **来源于第三方组件**

   框架还支持注入第三方组件，实现与外部系统的无缝集成：

   ```go
   type UseDep struct {
       gone.Flag
       redis *redis.Client `gone:"*"` // 需要注入的字段，来源于第三方组件
   }
   ```

   这种能力使Gone框架能够与各种外部库和服务协同工作，扩展应用程序的功能范围。

## Gone框架的启动流程

Gone框架提供了多种启动和管理应用的方式，让开发者可以根据不同的需求灵活选择。

### `gone.NewApp`方法

通过`gone.NewApp`可以创建一个`gone.Application`实例，这是应用程序的核心容器：

```go
app := gone.NewApp()
```

创建实例后，可以通过`Application::Load`和`Application::Loads`方法将Goner对象加载到框架中，为依赖注入做好准备。

### `gone.Application`的`Run`方法

在加载完所有Goner对象后，可以通过`Application::Run`方法启动框架：

```go
app.Run(func(service *MyService) {
    service.DoSomething()
})
```

`Run`方法支持传入多个函数作为参数，这些函数会按照顺序执行，并且支持函数参数注入。这种设计非常适合执行一系列初始化任务或启动多个并行服务。

### `gone.Application`的`Serve`方法

对于需要长时间运行的服务，Gone框架提供了`Serve`方法：

```go
app.Serve()
```

`Application::Serve`与`Application::Run`方法类似，但`Serve`方法会阻塞当前线程，直到服务接收到停止信号或调用`Application::End`来手动停止服务。这种方式特别适用于后台服务程序和Web应用。

需要注意的是，`Serve`方法不支持传入参数，因为它主要用于启动已经完全初始化的长期运行服务。

### `gone.Default`默认实例

为了简化使用，Gone框架提供了一个默认的`Application`实例，通过以下全局方法可以直接操作这个默认实例：

- `gone.Run` - 运行应用程序
- `gone.Serve` - 启动长期运行的服务
- `gone.End` - 停止服务
- `gone.Load` - 加载单个Goner对象
- `gone.Loads` - 加载多个Goner对象

这种设计使简单应用程序的编写变得更加简洁，无需显式创建应用实例。

## 将对象加载到Gone框架

Gone框架提供了多种灵活的方式来加载对象，使开发者能够根据项目的组织结构和依赖关系选择最合适的方法。

### `gone.Loader`和`gone.LoadFunc`

`gone.Loader`是Gone框架核心提供的用于加载对象的接口，它定义了如何将对象加载到框架中：

```go
type Loader interface {
    Load(goner Goner, options ...Option) error
    MustLoad(goner Goner, options ...Option) Loader // 加载goner，如果加载失败会panic，支持链式调用
    MustLoadX(x any) Loader // 加载x，x可以为Goner 或者 LoadFunc
    Loaded(LoaderKey) bool
}
```

而`gone.LoadFunc`是一个函数类型，定义了加载函数：

```go
type LoadFunc = func(Loader) error
```

这种设计允许我们将业务逻辑封装为组件，每个组件可能包含多个可注入的对象。通过将加载逻辑封装到`LoadFunc`函数中，可以方便地将相关对象统一加载到框架中：

```go
package componentA
import "github.com/gone-io/gone/v2"

type A struct {
    gone.Flag
}

func ALoad(loader gone.Loader) error {
    // 加载B组件相关依赖
    loader.MustLoadX(componentB.BLoad)
    
    // 加载A组件相关对象
    loader.MustLoad(&A{})

    return nil
}
```

更强大的是，这种设计支持组件之间的依赖关系：如果A组件依赖B组件，可以在A组件的`LoadFunc`函数中首先加载B组件，确保依赖项在被依赖之前就已经准备好。

### 多种加载方式的综合示例

下面通过一个综合示例，展示Gone框架提供的多种对象加载方式：

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
            // 1.使用`gone.NewApp`加载对象，支持多个`LoadFunc`函数作为参数
        ).
        // 2.使用`Application`的实例方法链式调用加载对象
        Load(&A{ID: "a"}, gone.Name("instance-a")).
        Load(&A{ID: "b"}, gone.Name("instance-b")).
        // 3.使用`Application`的实例`Loads`方法通过`LoadFunc`方法加载对象
        Loads(
            func(loader gone.Loader) error {
                // 4.使用`gone.Loader`链式调用加载对象
                loader.
                    MustLoad(&A{ID: "c"}, gone.Name("instance-c")).
                    MustLoad(&A{ID: "d"}, gone.Name("instance-d")).
                    // 5.通过`gone.Loader::MustLoadX`方法再调用其他的`LoadFunc`方法加载对象
                    MustLoadX(func(loader gone.Loader) error {
                        return loader.Load(&A{ID: "f"}, gone.Name("instance-f"))    
                    })
                
                return loader.Load(&A{ID: "e"}, gone.Name("instance-e"))
            },
            // 支持多个`LoadFunc`函数作为参数
        ).
        Run(func(a []*A) {
            fmt.Printf("%#v", a)
        })
        // 也可以使用Serve()启动长期运行的服务
}
```

这个例子展示了多种加载对象的方式，从简单的直接加载到复杂的链式和嵌套加载，满足不同的组织结构和依赖关系需求。

## 手动控制依赖注入

虽然Gone框架的自动依赖注入机制已经能够处理大多数场景，但有时我们可能需要在特定情况下手动控制依赖注入过程。为此，框架提供了两个专用接口：

### 结构体注入（StructInjector）

`gone.StructInjector`接口用于手动对结构体字段进行注入。这在运行时动态创建对象时特别有用：

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
    // "手动"完成结构体字段注入
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

在这个例子中，`Business`类获取了一个`StructInjector`，然后使用它来手动注入`User`结构体的字段。这种方式允许在运行时动态创建和注入对象，非常适合处理用户输入或配置驱动的场景。

### 函数参数注入（FuncInjector）

`gone.FuncInjector`接口用于手动对函数参数进行注入。这在处理回调、中间件或事件处理器时特别有用：

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
    
    // "手动"完成函数参数的注入
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

在这个例子中，`Business`类获取了一个`FuncInjector`，然后使用它来包装函数`needInjectedFunc`，自动注入函数参数。这种方式使处理回调和事件变得更加简洁和灵活。

## 总结

Gone框架提供了一套强大而灵活的依赖注入机制，使Go语言开发者能够构建松耦合、可测试和可维护的应用程序。通过结构体字段注入和函数参数注入，框架满足了各种复杂场景的需求。同时，框架提供的多种加载方式和手动控制机制，进一步增强了开发者的灵活性和控制力。

依赖注入是现代软件开发中的重要概念，它改变了我们组织和管理代码的方式。通过Gone框架的依赖注入机制，我们可以更轻松地构建大型、复杂的应用程序，而不必担心组件间的紧密耦合或难以测试的问题。

随着对Gone框架的深入学习和实践，您将能够充分利用依赖注入的优势，构建出更加健壮和可维护的Go应用程序。