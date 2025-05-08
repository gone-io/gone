<p>
    <a href="v2-update.md">English</a>&nbsp ｜&nbsp 中文
</p>

# Gone v2 使用说明

- [Gone v2 使用说明](#gone-v2-使用说明)
	- [Goner的定义](#goner的定义)
	- [Goner加载器](#goner加载器)
		- [加载Goner](#加载goner)
		- [加载函数](#加载函数)
		- [加载选项](#加载选项)
	- [依赖注入](#依赖注入)
		- [基于字段类型的注入](#基于字段类型的注入)
		- [基于Goner名称注入](#基于goner名称注入)
		- [通过Provider给组件注入依赖](#通过provider给组件注入依赖)
	- [Goner的生命周期](#goner的生命周期)
	- [四个hook函数](#四个hook函数)
	- [Application](#application)
	- [GonerKeeper](#gonerkeeper)
	- [数组注入](#数组注入)
	- [使用FuncInjector来实现函数参数的注入](#使用funcinjector来实现函数参数的注入)
	- [内置的`Config`和`Logger`](#内置的config和logger)
		- [Config](#config)
		- [Logger](#logger)


## Goner的定义

Goner是Gone框架定义的组件，是只嵌入了`gone.Flag`的结构体指针，可以用于Gone框架的依赖注入，如下代码就定义了一个简单的Goner：

```go
package demo

import "github.com/gone-io/gone"

type Component struct {
	gone.Flag
}

var aGoner = &Component{}

```

## Goner加载器

Gone框架核心是一个Goner仓库，加载器的作用就是将用户定义的组件（Goner）注册（或者加载，在本文后续我们都称为加载）到仓库中，以便后续的依赖注入。

### 加载Goner

```go
package main

import "github.com/gone-io/gone"

func main() {
	type Dep struct {
		gone.Flag
		Name string
	}

	// 加载一个Dep
	gone.Load(&Dep{})

	//加载 一个名为dep1的Dep
	gone.Load(&Dep{}, gone.Name("dep1"))

	//支持链式调用
	gone.
		Load(&Dep{}, gone.Name("dep3")).
		Load(&Dep{}, gone.Name("dep2"))

	//通过加载函数批量加载
	gone.Loads(func(loader gone.Loader) error {
		err := loader.Load(&Dep{}, gone.Name("dep4"))
		if err != nil {
			return gone.ToError(err)
		}
		err = loader.Load(&Dep{}, gone.Name("dep5"))

		return err
	})
}
```

上面的代码展示了Gone框架中加载Goner的几种方式：

1. 直接加载：使用`gone.Load()`方法可以直接加载一个Goner。例如`gone.Load(&Dep{})`将加载一个默认的Dep组件。

2. 命名加载：通过`gone.Name()`选项可以为加载的Goner指定一个名称。例如`gone.Load(&Dep{}, gone.Name("dep1"))`将加载一个名为"dep1"的Dep组件。

3. 链式加载：Gone框架支持链式调用方式加载多个Goner。可以通过`.Load()`方法连续加载多个组件，使代码更加简洁。

4. 批量加载：使用`gone.Loads()`方法可以在一个函数中批量加载多个Goner。这种方式特别适合需要进行错误处理的场景，可以统一处理加载过程中的错误。

这些加载方式提供了灵活的组件注册机制，开发者可以根据具体需求选择合适的方式来加载Goner。

### 加载函数

查看源代码，可以看到`gone.Loads()`方法的定义:

```go
func (s *Application) Loads(loads ...LoadFunc) *Application {
//...
}
```

`LoadFunc`是一个函数类型，定义如下：

```go
type LoadFunc func (Loader) error
```

如果编写一个功能模块，需要加载多个Goner到Gone框架中，可以提供一个`LoadFunc`函数，业务代码只需要通过`gone.Loads()`
方法调用这个函数即可。

### 加载选项

查到源代码，可以看到`gone.Load()`方法的定义:

```go
func Load(goner Goner, options ...Option) *Application {
//...
}
```

`Option`就是加载选项，它的作用是在加载Goner时，可以设置一些选项，比如`gone.Name("dep1")`就是设置Goner的名称为"dep1"。

**支持的选项**：

- `gone.IsDefault(objPointers ...any)`: 将组件标记为其类型的默认实现。当存在多个相同类型的组件时，如果没有指定具体名称，将使用默认实现进行注入。
  ```go
  // 将EnvConfigure标记为默认实现
  gone.Load(&EnvConfigure{}, gone.IsDefault())
  ```

- `gone.Order(order int)`: 设置组件的启动顺序。order值越小，组件启动越早。框架提供了三个预设的顺序选项：
    - `gone.HighStartPriority()`: 相当于`Order(-100)`，最早启动
    - `gone.MediumStartPriority()`: 相当于`Order(0)`，默认启动顺序
    - `gone.LowStartPriority()`: 相当于`Order(100)`，最后启动
  ```go
  // Database会在Service之前启动
  gone.Load(&Database{}, gone.Order(1))  // 先启动
  gone.Load(&Service{}, gone.Order(2))   // 后启动
  ```

- `gone.Name(name string)`: 为组件设置一个自定义名称。组件可以通过这个名称被注入。
  ```go
  // 加载一个名为"configure"的组件
  gone.Load(&EnvConfigure{}, gone.Name("configure"))
  ```

- `gone.OnlyForName()`: 标记组件仅支持基于名称的注入。使用此选项时，组件不会被注册为类型提供者，只能通过显式引用其名称进行注入。
  ```go
  // EnvConfigure只能通过`gone:"configure"`标签注入
  gone.Load(&EnvConfigure{}, gone.Name("configure"), gone.OnlyForName())
  ```

- `gone.ForceReplace()`: 允许替换具有相同名称或类型的现有组件。加载带有此选项的组件时：
    - 如果存在同名组件，将被替换
    - 如果存在相同类型的提供者，将被替换
  ```go
  // 这将替换任何名为"service"的现有组件
  gone.Load(&MyService{}, gone.Name("service"), gone.ForceReplace())
  ```

- `gone.LazyFill()`: 将组件标记为延迟填充。使用此选项时，组件只有在实际被注入时才会被加载。这对于加载成本高或有外部依赖的组件很有用。
  ```go
  // 只有在实际注入时才会加载组件
  gone.Load(&MyService{}, gone.Name("service"), gone.LazyFill())
  ```

## 依赖注入

### 基于字段类型的注入

```go
package main

import "github.com/gone-io/gone"

type Dep struct {
	gone.Flag
	Name string
}
type Service struct {
	gone.Flag
	dep1 *Dep `gone:""` // 默认注入
	dep2 *Dep `gone:"*"`
}
```

在上面的代码中展示了基于类型的注入方式：
按类型注入，需要在被注入的的字段上添加`gone:""`(或者 `gone:"*"`，在v2中两者是等效的)
标签，框架会自动根据类型进行注入。如果在注入前，加载了多个相同类型的组件，那么框架会优先选择默认实现（通过`IsDefault()`
选项设置），否则会选择第一个加载的组件并提示警告。

### 基于Goner名称注入

gone标签中可以指定组件名称，框架会根据名称进行注入。

```go
package main

import "github.com/gone-io/gone"

type Dep struct {
	gone.Flag
	Name string
}
type Service struct {
	gone.Flag
	dep1 *Dep `gone:"dep1"` // 指定名称注入
	dep2 *Dep `gone:"dep2"` // 指定名称注入
}
```

Goner支持两种方式设置名称：

1. 在加载时，使用`gone.Name()`选项设置名称。
2. Goner实现了`NamedGoner`接口，通过`GonerName`方法设置名称，NamedGoner接口定义如下：

```go
type NamedGoner interface {
	Goner
	GonerName() string
}
```

### 通过Provider给组件注入依赖

在v1中，给组件注入配置参数是这样写的：

```go
type Service struct {
gone.Flag
confStr string `gone:"config,configKeyName"` // 通过标签注入配置参数
}
```

在v2中，依然支持这样方式，底层实现改为由Provider提供值。
让我们来看看Provider的接口定义：

```go
type Provider[T any] interface {
	Goner
	Provide(tagConf string) (T, error)
}
```

它是一个泛型接口，在实际使用中，我们需要定义一个Provider，并实现`Provide`方法，在`Provide`
方法中，我们可以通过tagConf获取到配置参数，然后返回值即可。必然为了支持上面的`confStr`的配置，我们需要定义一个Provider，如下：

```go
type ConfigProvider struct {
gone.Flag
}

func (c *ConfigProvider) Provide(tagConf string) (string, error) {
return config.Get(tagConf)
}
```

在`Provide`方法中，我们通过`tagConf`获取到配置参数，然后返回值即可。这样，我们就可以通过Provider给组件注入配置参数了。

如果需要被配置的字段是一个int类型，那么我们需要定义一个Provider，如下：

```go
type ConfigProvider struct {
	gone.Flag
}

func (c *ConfigProvider) Provide(tagConf string) (int, error) {
	return config.Get(tagConf)
}
```

会很快发现一个问题，为了实现一个Config模块，我们需要定义无数个Provider，这显然是不合理的，因此，我们需要一个通用的Provider，如下：

```go
type NamedProvider interface {
	NamedGoner
	Provide(tagConf string, t reflect.Type) (any, error)
}
```

NamedProvider 接口定义了一个`Provide`
方法，该方法接收两个参数，第一个参数是tagConf，第二个参数是t，t是reflect.Type类型，用于获取字段的类型，然后根据字段的类型返回对应的值。这样，我们就可以通过NamedProvider给组件注入配置参数了。

现在对比一下 Provider 和 NamedProvider 的区别：

1. Provider 是一个泛型，它可以根据字段的类型返回对应的值，我们每次实现的Provider都只能固定一个类型，所以只能返回一个值，不能返回多个类型。它的应用场景是，是按类型注入第三方的值。
2. NamedProvider 是一个接口，它可以根据字段的类型返回对应的值。它的应用场景是，需要通过一个Provider给组件注入多个类型的值。

下面讲讲依赖注入查找的流程：

1. 如果gone标签中没有指定名称或者指定的名称为`*`，那么框架会使用内核中提供Provider来给组件注入值（对，内核其实就是一个Provider，v2的整个注入机制都是基于Provider的）。

- 内核Provider会按类型查找Provider，如果找到了，那么就会调用Provider的`Provide`方法，将值注入到组件中。
- 如果没有找到，内核Provider会按类型查找加载到Goner仓库的值，如果能找到兼容的值，那么就会将值注入到组件中。
- 如果还是没有找到，那么就会报错。

2. 如果gone标签中指定了名称，那么框架会使用指定的Provider来给组件注入值。

- 优先按名字查找`Provider[T any]` 和 `NoneParamProvider[T any]`，如果找到了，那么就会调用Provider的`Provide`
  方法，如果他们提供的值是兼容的，那么就会将值注入到组件中。
- 如果还没有注入成功，则继续按名字查找`NamedProvider`，如果找到了，那么就会调用它的
  `Provide(tagConf string, t reflect.Type) (any, error)`方法，它如果能返回一个兼容的值，那么就会将值注入到组件中。
- 如果还是没能注入成功，还用回调`StructFieldInjector`来注入值。
- 如果还是没有找到，那么就会报错。

> 补充说明：
> NoneParamProvider的定义：
> ```go
> type NoneParamProvider[T any] interface {
>   Goner
> 	Provide() T
> }
> ```
> 它是一个泛型接口，它只有一个方法`Provide`，这个方法没有参数，返回一个值。它的应用场景是，需要通过一个Provider给组件注入一个值，这个值不需要通过参数传递，只需要通过方法返回即可。

> StructFieldInjector的定义：
> ```go
> type StructFieldInjector interface {
>   NamedGoner
> 	Inject(tagConf string, field reflect.StructField, fieldValue reflect.Value) error
> }
> ```
> 它是一个接口，它只有一个方法`InjectStructField`，这个方法接收三个参数，第一个参数是v，第二个参数是tagConf，第三个参数是t，t是reflect.Type类型，用于获取字段的类型，然后根据字段的类型返回对应的值。它的应用场景是，需要通过一个Provider给组件注入多个类型的值。

从接口的定义上可以看到，Provider、NoneParamProvider、NamedProvider和StructFieldInjector都是Goner的子接口，要实现他们都必须嵌入
`gone.Flag`。他们的用途都是将第三方的值提给Gone框架，让框架来进行依赖注入。

## Goner的生命周期

![flow.png](assert/flow.png)

1. 初始化阶段
   依赖注入前，如果组件上存在 `BeforeInit() error` 或者 `BeforeInit() `，那么就会调用这个方法，这个方法会在依赖注入之前被调用，
   **在这个方法中不能够使用依赖注入的值**。
   依赖注入后，如果组件上存在 `Init() error` 或者 `Init() `，那么就会调用这个方法，这个方法会在依赖注入之后被调用，*
   *在这个方法中可以使用依赖注入的值**。
   组件被注入到其他组件前，必须先完成自己的初始化。
2. 运行阶段
   这个阶段，如果组件实现了`Daemon`接口，该阶段会运行`Daemon`接口的`Start() error`方法来启动自己。可以在加载组件时，通过
   `gone.Order()`方法设置组件的启动顺序。
3. 停机阶段
   这个阶段，如果组件实现了`Daemon`接口，该阶段会运行`Daemon`接口的`Stop() error`方法来停止自己。`Stop`的顺序和`Start`
   的顺序相反。

注意：如果需要Daemon持续提供服务，需要调用`Serve()`方法，而不是`Run()`，Serve函数会阻塞，直到`End`被调用或者进程收到终止的信号。

```go
package use_case

import (
	"github.com/gone-io/gone"
	"testing"
	"time"
)

type testDaemon struct {
	gone.Flag
	isStart bool
	isStop  bool
}

func (t *testDaemon) Start() error {
	println("testDaemon Start")
	t.isStart = true
	return nil
}

func (t *testDaemon) Stop() error {
	println("testDaemon Stop")
	t.isStop = true
	return nil
}

func TestServe(t *testing.T) {
	daemon := &testDaemon{}
	var t1, t2 time.Time
	ins := gone.NewApp() //这里创建了一个实例，在后面的##Application章节有说明。 

	ins.
		Load(daemon).
		BeforeStart(func() {
			go func() {
				time.Sleep(5 * time.Millisecond)
				t1 = time.Now()
				ins.End() //如果调用gone.End()方法，将终止框架默认的实例`gone.Default`
			}()
		}).
		AfterStop(func() {
			t2 = time.Now()
		}).
		Serve()

	if !daemon.isStart {
		t.Fatal("daemon start failed")
	}
	if !daemon.isStop {
		t.Fatal("daemon stop failed")
	}
	if !t2.After(t1) {
		t.Fatal("daemon stop after serve failed")
	}
}
```

## 四个hook函数

框架提供4个hook函数，分别为作用于beforeStart、afterStart、beforeStop、afterStop。其中两个before Hook 先注册的后执行；两个after
Hook 先注册的先执行。下面代码是通过依赖注入的方式来演示的，只能在组件完成初始化之后才能使用。

```go
package use_case

import (
	"github.com/gone-io/gone"
	"testing"
)

type hookTest struct {
	gone.Flag
	beforeStart gone.BeforeStart `gone:""`
	afterStart  gone.AfterStart  `gone:""`
	beforeStop  gone.BeforeStop  `gone:""`
	afterStop   gone.AfterStop   `gone:""`
}

var orders []int

func (h *hookTest) Init() {
	//通过注入的BeforeStart方法，可以注册一个函数，在服务启动前执行
	h.beforeStart(func() {
		println("before start 1")
		orders = append(orders, 1)
	})

	//before 类型的hook 可以注册多个，先注册的后执行，后注册的先执行
	h.beforeStart(func() {
		println("before start 2")
		orders = append(orders, 2)
	})

	h.afterStart(func() {
		println("after start 3")
		orders = append(orders, 3)
	})

	h.afterStart(func() {
		println("after start 4")
		orders = append(orders, 4)
	})

	h.beforeStop(func() {
		println("before stop 5")
		orders = append(orders, 5)
	})

	h.beforeStop(func() {
		println("before stop 6")
		orders = append(orders, 6)
	})

	h.afterStop(func() {
		println("after stop 7")
		orders = append(orders, 7)
	})
	h.afterStop(func() {
		println("after stop 8")
		orders = append(orders, 8)
	})
}

func TestUseHook(t *testing.T) {
	gone.Load(&hookTest{}).Run()

	wantedOrder := []int{2, 1, 3, 4, 6, 5, 7, 8}
	for i := range wantedOrder {
		if wantedOrder[i] != orders[i] {
			t.Errorf("wanted %v, got %v", wantedOrder[i], orders[i])
		}
	}
}
```

hook函数也可以直接在加载Goner组件后注册，像这样：

```go
func TestUseHookDirectly(t *testing.T) {
    type testGoner struct {
        gone.Flag
    }
    gone.
        Load(&testGoner{}).
        // 直接注册Hook函数
        BeforeStart(func () {
            println(" BeforeStart")
        }).
        AfterStart(func () {
            println(" AfterStart")
        }).
        Run()
}
```

## Application

查看源代码，可以发现：`gone.Load`、`gone.Loads`、`gone.Run`、`gone.Serve`等函数实际是调用的`Application`的一个实例`Default`
上的对应方法。

```go
var Default = NewApp()
//...
func Load(goner Goner, options ...Option) *Application {
    return Default.Load(goner, options...)
}
//...
func Loads(loads ...LoadFunc) *Application {
    return Default.Loads(loads...)
}
//...
```

所以，如果希望在同一个进程中使用多个Gone框架实例，可以使用 `gone.NewApp` 函数来创建多个 `Application` 实例，然后分别调用 `Run` 或`Serve` 方法启动框架。
下面是`NewApp`的定义：
```go
func NewApp(loads ...LoadFunc) *Application {
    application := Application{}
    //....
    return &application
}
```

## GonerKeeper

如果希望在组件中，使用代码动态地获取其他组件，可以注入`gone.GonerKeeper`接口（当然也可以直接注入 *gone.Core），这个接口定义如下：

```go
type GonerKeeper interface {
    GetGonerByName(name string) any
    GetGonerByType(t reflect.Type) any
}
```
示例代码：
```go
package use_case

import (
	"github.com/gone-io/gone"
	"testing"
)

type useKeeper struct {
	gone.Flag
	keeper gone.GonerKeeper `gone:"*"`
	core   *gone.Core       `gone:"*"`
}

func (u *useKeeper) Test(t *testing.T) {
	goner := u.keeper.GetGonerByName("*")
	if goner != u.core {
		t.Fatal("keeper get core error")
	}
}

func TestGonerKeeper(t *testing.T) {
	gone.
		NewApp().
		Load(&useKeeper{}).
		Run(func(k *useKeeper) {
			k.Test(t)
		})
}
```

## 数组注入
在v2版本中，依然可以使用 接口slice 来接收 多个实例。
测试代码如下：
```go
package use_case

import (
	"github.com/gone-io/gone"
	"testing"
)

type worker interface {
	Work()
}

type workerImpl struct {
	gone.Flag
	name string
}

func (w *workerImpl) Work() {
	println("worker", w.name, "work")
}

type workerImpl2 struct {
	gone.Flag
	name string
}

func (w *workerImpl2) Work() {
	println("worker", w.name, "work")
}

type factory struct {
	gone.Flag
	workers []worker `gone:"*"`
}

func TestUseSlice(t *testing.T) {
	gone.
		NewApp().
		Load(&factory{}, gone.Name("factory")).
		Load(&workerImpl{name: "worker1"}, gone.Name("worker1")).
		Load(&workerImpl2{name: "worker2"}, gone.Name("worker2")).
		Run(func(f *factory) {
			if len(f.workers) != 2 {
				t.Fatal("worker count is not 2")
			}
		})

}
```
## 使用FuncInjector来实现函数参数的注入
函数参数注入是Gone框架的一个特性，它允许框架在调用函数时，自动从Goner仓库中查找并注入与函数参数类型匹配的组件实例。
在前面例子中，Run方法接收的函数，其参数就是被自动注入的。
例子：
```go
package use_case

import (
	"github.com/gone-io/gone"
	"testing"
)

type funcTest struct {
	gone.Flag
	injector gone.FuncInjector `gone:"*"`
}

func (f *funcTest) Test(t *testing.T) {
	// 定义一个需要注入参数的函数
	fn := func(factory *factory) {
		if factory == nil {
			t.Fatal("factory is nil")
		}
	}

	// 使用 InjectWrapFunc 来执行函数，框架会自动注入参数
	wrapped, err := f.injector.InjectWrapFunc(fn, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = wrapped()

	// 也可以注入多个参数
	fn2 := func(factory *factory, worker worker) {
		if factory == nil {
			t.Fatal("factory is nil")
		}
		if worker == nil {
			t.Fatal("worker is nil")
		}
	}

	wrapped2, err := f.injector.InjectWrapFunc(fn2, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = wrapped2()
}

func TestFuncInjector(t *testing.T) {
	gone.
		NewApp().
		Load(&funcTest{}).
		Load(&factory{}, gone.Name("factory")).
		Load(&workerImpl{name: "worker1"}, gone.Name("worker1")).
		Run(func(f *funcTest) {
			f.Test(t)
		})
}
```
## 内置的`Config`和`Logger`

在v2版本中，内核代码内置了`Config`和`Logger`组件，分别用于配置管理和日志记录。

### Config
内置的`Config`组件，是从环境变量中读取配置，可以通过实现`gone.Configure`来实现自定义读取配置的方式。
下面结束如何读取配置的示例代码，注意环境变量的名需要加上`GONE_`前缀，并且需要全部大写，如果被注入的不是简单类型，默认的Configure会尝试使用json解析环境变量的值。
```go
package use_case

import (
	"github.com/gone-io/gone/v2"
	"os"
	"testing"
)

type useConfig struct {
	gone.Flag
	goneVersion string `gone:"config,gone-version"` //框架在启动时，会自动加载配置，并注入到goRoot字段中。
}

func TestUseConfig(t *testing.T) {
	os.Setenv("GONE_GONE-VERSION", "v2.0.0")

	gone.
		Load(&useConfig{}).
		Run(func(c *useConfig) {
			println("goRoot:", c.goneVersion)
			if c.goneVersion != "v2.0.0" {
				t.Fatal("配置注入失败")
			}
		})
}
```


### Logger
v2内核内置的Logger，只是简单在控制台打印日志，可以通过实现`gone.Logger`来实现自定义日志记录的方式。
```go
package use_case

import (
	"github.com/gone-io/gone/v2"
	"testing"
)

type worker struct {
	gone.Flag
	log gone.Logger `gone:"*"`
}

func TestUseLogger(t *testing.T) {
	gone.
		Load(&worker{}).
		Run(func(app *app) {
			app.log.Infof("hello world")
			app.log.Errorf("hello world")
			app.log.Warnf("hello world")
			app.log.Debugf("hello world")
		})
}
```


