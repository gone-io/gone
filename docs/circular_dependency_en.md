# Circular Dependency Issues in Goner

- [Circular Dependency Issues in Goner](#circular-dependency-issues-in-goner)
  - [Examples](#examples)
  - [Circular Dependency Analysis](#circular-dependency-analysis)
    - [Initialization Flow in the Gone Framework](#initialization-flow-in-the-gone-framework)
    - [Component Types and Dependency Collection](#component-types-and-dependency-collection)
    - [Dependency Collection and Circular Dependency Detection](#dependency-collection-and-circular-dependency-detection)
  - [When Does a Circular Dependency Panic Occur?](#when-does-a-circular-dependency-panic-occur)
    - [Analysis of the Differences Between the Test Cases](#analysis-of-the-differences-between-the-test-cases)
  - [Why Have Circular Dependency Panic?](#why-have-circular-dependency-panic)
  - [How to Solve Circular Dependency Problems?](#how-to-solve-circular-dependency-problems)
    - [Example: Solving the Circular Dependency in Case 2](#example-solving-the-circular-dependency-in-case-2)


## Examples
Let's look at several test cases:

- Case 1:
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

- Case 2:
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

- Case 3:
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

- Case 4:
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

Test results:
- Case 1: Normal execution
- Case 2: Panic, output:
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
- Case 3: Normal execution
- Case 4: Normal execution

In all four cases, two structs depend on each other, but only "Case 2" throws a circular dependency error. Why?

## Circular Dependency Analysis

### Initialization Flow in the Gone Framework

The initialization flow in the Gone framework is mainly divided into two phases:

1. **fillAction (Field Filling)**: The framework injects dependencies into component fields
2. **initAction (Component Initialization)**: The framework calls the component's `Init()` method for initialization

These two phases are represented internally in the framework as different `actionType` constants:
```go
const (
	fillAction actionType = 1
	initAction actionType = 2
)
```

### Component Types and Dependency Collection

Components in the Gone framework can be categorized into several types, including:

1. **Regular Goner**: Structures that only embed `gone.Flag` without implementing special interfaces
   - Only require field filling (fillAction)
   - Not marked as `needInitBeforeUse`

2. **Init Goner**: Components that implement the `Initiator` or `InitiatorNoError` interface (have an `Init()` method)
   - Require both field filling and initialization (fillAction and initAction)
   - Marked as `needInitBeforeUse`

When creating a component, the framework checks if the component implements specific interfaces to determine if it needs to be initialized before use:

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

### Dependency Collection and Circular Dependency Detection

During the dependency collection process, the framework collects two types of dependencies for each component:

1. **fillDependency**: Dependencies required for field filling of the component
2. **initDependency**: Dependencies required for component initialization (usually fillAction dependencies)

The key is that when a field depends on a component marked as `needInitBeforeUse`, the framework adds an additional initAction dependency:

```go
if depCo.needInitBeforeUse {
    fillDependencies = append(fillDependencies, dependency{
        coffin: depCo,
        action: initAction,
    })
}
```

This means that if component A depends on component B, and B needs initialization, then A's field filling depends not only on B's existence but also on B's initialization completion.

## When Does a Circular Dependency Panic Occur?

Based on the above analysis, the Gone framework will detect a circular dependency and throw a panic when the following conditions are met:

1. **Two or more components have mutual dependency relationships** (A depends on B, B depends on A)
2. **These components all implement the `Init()` method**, marked as `needInitBeforeUse`

In this case, the dependency relationship forms an unresolvable cycle:
- A's field filling depends on B's initialization
- B's initialization depends on B's field filling
- B's field filling depends on A's initialization
- A's initialization depends on A's field filling

This forms an unbreakable circular dependency chain, and the framework cannot determine the initialization order, so it throws a panic.

### Analysis of the Differences Between the Test Cases

Now we can explain the different behaviors of the test cases:

1. **Case 1**: depA1 and depB1 are both Regular Goners
   - Neither component implements the `Init()` method
   - Neither is marked as `needInitBeforeUse`
   - Only have fillAction dependencies, no initAction dependencies
   - Although there is circular referencing, the framework allows this type of cycle because it only involves field filling, not initialization order

2. **Case 2**: dep1 and dep2 are both Init Goners
   - Both components implement the `Init()` method
   - Both are marked as `needInitBeforeUse`
   - Have both fillAction and initAction dependencies
   - The dependency relationship forms a true cycle:
     - dep1.fill → dep2.init → dep2.fill → dep1.init → dep1.fill
   - The framework cannot determine the initialization order, so it reports an error

3. **Case 3**: depA3 is a Regular Goner, depB3 is an Init Goner
   - depA3 does not implement the `Init()` method, not marked as `needInitBeforeUse`
   - depB3 implements the `Init()` method, marked as `needInitBeforeUse`
   - The dependency relationship does not form a complete cycle:
     - depA3.fill → depB3.init → depB3.fill
   - Since depA3 only has fillAction, no initAction, a complete circular dependency is not formed

## Why Have Circular Dependency Panic?

The Gone framework designed the circular dependency detection mechanism mainly to solve the following problems:

1. **Ensure deterministic initialization order**: If circular dependencies exist, the framework cannot determine the initialization order of components, which may cause some components to be used before their dependencies are fully initialized

2. **Prevent initialization deadlock**: Circular dependencies can cause the initialization process to deadlock, especially when all components need initialization

3. **Early detection of design problems**: Circular dependencies usually indicate problems in application design; early detection and reporting can help developers improve the design

The framework uses a topological sorting algorithm to detect cycles in the dependency graph, and if a cycle is found, it throws a panic to prompt developers to solve this problem.

## How to Solve Circular Dependency Problems?

When encountering circular dependency problems, the following methods can be used to solve them:

1. **Refactor component design**:
   - Re-examine the dependency relationships between components, consider whether the design can be restructured to eliminate circular dependencies
   - May need to introduce new abstraction layers or intermediate components to break the cycle

2. **Use interfaces for decoupling**:
   - Change direct dependencies to interface dependencies, then have both components implement the same interface
   - This can reduce direct coupling between components

3. **Use Regular Goner**:
   - If possible, change one of the components to a Regular Goner (not implementing the `Init()` method)
   - As shown in Case 3, this can avoid forming a complete circular dependency

4. **Use delayed initialization**:
   - Use the `LazyFill()` option to load Goner, delaying Goner's assembly (`fillAction`)
   - **Please note**: Side effects of using the `LazyFill()` option to load Goner:
      a. For the delayed Goner, dependency-injected fields cannot be used in methods named `Init`, `Provide`, `Inject`

5. **Use `option:"lazy"` to delay field injection**
   - Use the `option:"lazy"` option to delay field injection
   - **Please note**: Fields marked with `option:"lazy"` cannot be used in methods named `Init`, `Provide`, `Inject`

6. **Use event mechanism**:
   - Implement indirect communication between components through event or message mechanisms
   - This can avoid direct circular references

7. **Use third-party components**:
   - Introduce an intermediate component, making the two originally mutually dependent components both depend on this intermediate component
   - The intermediate component can hold necessary state or provide necessary services

### Example: Solving the Circular Dependency in Case 2

Here are several methods to solve the circular dependency in Case 2:

1. **Method 1: Use Regular Goner**
```go
type dep1 struct {
    gone.Flag
    dep *dep2 `gone:"*"`
}

// Remove Init method

type dep2 struct {
    gone.Flag
    dep *dep1 `gone:"*"`
}

func (d *dep2) Init() {}
```

2. **Method 2: Use interfaces for decoupling**
```go
type Dep2Interface interface {
    // Define necessary methods
    SomeMethod() error
}

type dep1 struct {
    gone.Flag
    dep Dep2Interface `gone:"*"` // Depend on interface rather than concrete implementation
}

func (d *dep1) Init() {}

type dep2 struct {
    gone.Flag
    // No longer directly depends on dep1
}

func (d *dep2) Init() {}
func (d *dep2) SomeMethod() error { return nil } // Implement interface
```

3. **Method 3: Use intermediate component**
```go
type mediator struct {
    gone.Flag
    dep1 *dep1 `gone:"*"`
    dep2 *dep2 `gone:"*"`
}

func (m *mediator) Init() {
    // Coordinate interaction between dep1 and dep2 here
}

type dep1 struct {
    gone.Flag
    // No longer directly depends on dep2
}

func (d *dep1) Init() {}

type dep2 struct {
    gone.Flag
    // No longer directly depends on dep1
}

func (d *dep2) Init() {}
```