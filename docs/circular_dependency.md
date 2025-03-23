# Goner的循环依赖问题

- [Goner的循环依赖问题](#goner的循环依赖问题)
  - [引例](#引例)
  - [循环依赖分析](#循环依赖分析)
    - [Gone框架的初始化流程](#gone框架的初始化流程)
    - [组件类型与依赖收集](#组件类型与依赖收集)
    - [依赖收集与循环依赖检测](#依赖收集与循环依赖检测)
  - [什么情况下会导致循环依赖的panic？](#什么情况下会导致循环依赖的panic)
    - [三个用例的区别分析](#三个用例的区别分析)
  - [为什么会有循环依赖的panic？](#为什么会有循环依赖的panic)
  - [如何解决循环依赖问题？](#如何解决循环依赖问题)
    - [示例：解决用例2的循环依赖](#示例解决用例2的循环依赖)


## 引例
先来看两个测试用例：

- 用例1：
```go
type depA1 struct {
	gone.Flag
	dep *depB1 `gone:"*"`
}

type depB1 struct {
	gone.Flag
	dep *depA1 `gone:"*"`
}

func TestCircularDependency1(t *testing.T) {
	gone.
		NewApp().
		Load(&depA1{}).
		Load(&depB1{}).
		Run()
}
```

- 用例2：
```go

type dep1 struct {
	gone.Flag
	dep *dep2 `gone:"*"`
}

func (d *dep1) Init() {}

type dep2 struct {
	gone.Flag
	dep *dep1 `gone:"*"`
}

func (d *dep2) Init() {}

func TestCircularDependency2(t *testing.T) {
	gone.
		NewApp().
		Load(&dep1{}).
		Load(&dep2{}).
		Run()
}
```

- 用例3：
```go
type depA3 struct {
	gone.Flag
	dep *depB3 `gone:"*"`
}
type depB3 struct {
	gone.Flag
	dep *depA3 `gone:"*"`
}

func (d *depB3) Init() {}

func TestCircularDependency3(t *testing.T) {
	gone.
		NewApp().
		Load(&depA3{}).
		Load(&depB3{}).
		Run()
}
```

- 用例4：
```go
type depA4 struct {
	gone.Flag
	dep *depB4 `gone:"*"`
}

func (d *depA4) Init() {}

type depB4 struct {
	gone.Flag
	dep *depA4 `gone:"*" option:"lazy"`
}

func (d *depB4) Init() {}

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

测试结果：
- 用例1，正常；
- 用例2，panic，输出：
```log
=== RUN   TestCircularDependency2
--- FAIL: TestCircularDependency2 (0.00s)
panic: GoneError(code=1003); circular dependency:
			<fill fields of "*github.com/gone-io/gone/v2/use_case.dep1"> depend on
				<initialize of "*github.com/gone-io/gone/v2/use_case.dep2"> depend on
					<fill fields of "*github.com/gone-io/gone/v2/use_case.dep2"> depend on
						<initialize of "*github.com/gone-io/gone/v2/use_case.dep1"> depend on
							<fill fields of "*github.com/gone-io/gone/v2/use_case.dep1">

	github.com/gone-io/gone/v2.circularDepsError({0x1400017c280?, 0x14000117a40?, 0x102c3e45a?})
		/Users/jim/works/gone-io/gone/dep.go:21 +0x1e8
	github.com/gone-io/gone/v2.(*Core).Check(0x14000138730)
		/Users/jim/works/gone-io/gone/core.go:204 +0x70
	...
```
- 用例3，正常；
- 用例4，正常；


在4个用例中，都是两个struct相互依赖，但是却只有"用例2"抛出了循环依赖的错误， 为什么呢？

## 循环依赖分析

### Gone框架的初始化流程

Gone框架的初始化流程主要分为两个阶段：

1. **fillAction（字段填充）**：框架将依赖注入到组件的字段中
2. **initAction（组件初始化）**：框架调用组件的`Init()`方法进行初始化

这两个阶段在框架内部被表示为不同的`actionType`常量：
```go
const (
	fillAction actionType = 1
	initAction actionType = 2
)
```

### 组件类型与依赖收集

Gone框架中的组件可以分为几种类型，其中与我们的例子相关的有：

1. **普通Goner**：只嵌入了`gone.Flag`的结构体，没有实现特殊接口
   - 只需要进行字段填充（fillAction）
   - 不会被标记为`needInitBeforeUse`

2. **Init Goner**：实现了`Initiator`或`InitiatorNoError`接口的组件（有`Init()`方法）
   - 需要进行字段填充和初始化（fillAction和initAction）
   - 会被标记为`needInitBeforeUse`

框架在创建组件时，会检查组件是否实现了特定接口来决定是否需要在使用前初始化：

```go
func newCoffin(goner any) *coffin {
    _, needInitBeforeUse := goner.(Initiator)
    if !needInitBeforeUse {
        _, needInitBeforeUse = goner.(InitiatorNoError)
    }
    // ...
    return &coffin{
        goner:             goner,
        defaultTypeMap:    make(map[reflect.Type]bool),
        needInitBeforeUse: needInitBeforeUse,
    }
}
```

### 依赖收集与循环依赖检测

框架在依赖收集过程中，会为每个组件收集两种依赖：

1. **fillDependency**：组件字段填充所需的依赖
2. **initDependency**：组件初始化所需的依赖（通常是fillAction依赖）

关键在于，当一个字段依赖的组件被标记为`needInitBeforeUse`时，框架会添加一个额外的initAction依赖：

```go
if depCo.needInitBeforeUse {
    fillDependencies = append(fillDependencies, dependency{
        coffin: depCo,
        action: initAction,
    })
}
```

这意味着，如果组件A依赖组件B，且B需要初始化，那么A的字段填充不仅依赖B的存在，还依赖B的初始化完成。

## 什么情况下会导致循环依赖的panic？

根据上述分析，当满足以下条件时，Gone框架会检测到循环依赖并抛出panic：

1. **两个或多个组件之间存在相互依赖关系**（A依赖B，B依赖A）
2. **这些组件都实现了`Init()`方法**，被标记为`needInitBeforeUse`

这种情况下，依赖关系会形成一个无法解决的循环：
- A的字段填充依赖B的初始化
- B的初始化依赖B的字段填充
- B的字段填充依赖A的初始化
- A的初始化依赖A的字段填充

这就形成了一个无法打破的循环依赖链，框架无法确定初始化顺序，因此会抛出panic。

### 三个用例的区别分析

现在我们可以解释三个用例的不同表现：

1. **用例1**：depA1和depB1都是普通Goner
   - 两个组件都没有实现`Init()`方法
   - 都不会被标记为`needInitBeforeUse`
   - 只有fillAction依赖，没有initAction依赖
   - 虽然有循环引用，但框架允许这种循环，因为它只涉及字段填充，不涉及初始化顺序

2. **用例2**：dep1和dep2都是Init Goner
   - 两个组件都实现了`Init()`方法
   - 都被标记为`needInitBeforeUse`
   - 既有fillAction依赖，也有initAction依赖
   - 依赖关系形成了真正的循环：
     - dep1.fill → dep2.init → dep2.fill → dep1.init → dep1.fill
   - 框架无法确定初始化顺序，因此报错

3. **用例3**：depA3是普通Goner，depB3是Init Goner
   - depA3没有实现`Init()`方法，不会被标记为`needInitBeforeUse`
   - depB3实现了`Init()`方法，被标记为`needInitBeforeUse`
   - 依赖关系不会形成完整的循环：
     - depA3.fill → depB3.init → depB3.fill
   - 由于depA3只有fillAction，没有initAction，所以不会形成完整的循环依赖

## 为什么会有循环依赖的panic？

Gone框架设计了循环依赖检测机制，主要是为了解决以下问题：

1. **确保初始化顺序的确定性**：如果存在循环依赖，框架无法确定组件的初始化顺序，可能导致某些组件在依赖未完全初始化的情况下被使用

2. **防止初始化死锁**：循环依赖可能导致初始化过程陷入死锁，特别是当所有组件都需要初始化时

3. **提前发现设计问题**：循环依赖通常表明应用程序的设计存在问题，提前检测并报错可以帮助开发者改进设计

框架通过拓扑排序算法检测依赖图中的循环，如果发现循环，就会抛出panic，提示开发者解决这个问题。

## 如何解决循环依赖问题？

当遇到循环依赖问题时，可以采用以下几种方法解决：

1. **重构组件设计**：
   - 重新审视组件之间的依赖关系，考虑是否可以重新设计以消除循环依赖
   - 可能需要引入新的抽象层或中间组件来打破循环

2. **使用接口解耦**：
   - 将直接依赖改为依赖接口，然后让两个组件都实现相同的接口
   - 这样可以降低组件之间的直接耦合

3. **使用普通Goner**：
   - 如果可能，将其中一个组件改为普通Goner（不实现`Init()`方法）
   - 如用例3所示，这样可以避免形成完整的循环依赖

4. **使用延迟初始化**：
   - 使用`LazyFill()`选项加载Goner， 延迟Goner的装配(`fillAction`)
   - **请注意**：使用`LazyFill()`选项加载Goner的副作用：
      a. 被延迟的Goner，在名为`Init`、`Provide`、`Inject`这些方法中，无法使用依赖注入的字段

5. **使用`option:"lazy"`延迟字段注入**
   - 使用`option:"lazy"`选项，延迟字段注入
   - **请注意**：使用`option:"lazy"`标记的字段，不能在名为`Init`、`Provide`、`Inject`的这些方法中使用；

6. **使用事件机制**：
   - 通过事件或消息机制实现组件间的间接通信
   - 这样可以避免直接的循环引用

7. **使用第三方组件**：
   - 引入一个中间组件，让原本相互依赖的两个组件都依赖这个中间组件
   - 中间组件可以持有必要的状态或提供必要的服务

### 示例：解决用例2的循环依赖

以下是几种解决用例2循环依赖的方法：

1. **方法一：使用普通Goner**
```go
type dep1 struct {
    gone.Flag
    dep *dep2 `gone:"*"`
}

// 移除Init方法

type dep2 struct {
    gone.Flag
    dep *dep1 `gone:"*"`
}

func (d *dep2) Init() {}
```

2. **方法二：使用接口解耦**
```go
type Dep2Interface interface {
    // 定义必要的方法
    SomeMethod() error
}

type dep1 struct {
    gone.Flag
    dep Dep2Interface `gone:"*"` // 依赖接口而非具体实现
}

func (d *dep1) Init() {}

type dep2 struct {
    gone.Flag
    // 不再直接依赖dep1
}

func (d *dep2) Init() {}
func (d *dep2) SomeMethod() error { return nil } // 实现接口
```

3. **方法三：使用中间组件**
```go
type mediator struct {
    gone.Flag
    dep1 *dep1 `gone:"*"`
    dep2 *dep2 `gone:"*"`
}

func (m *mediator) Init() {
    // 在这里协调dep1和dep2的交互
}

type dep1 struct {
    gone.Flag
    // 不再直接依赖dep2
}

func (d *dep1) Init() {}

type dep2 struct {
    gone.Flag
    // 不再直接依赖dep1
}

func (d *dep2) Init() {}
```
