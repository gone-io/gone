<p>
   English&nbsp ｜&nbsp <a href="gone-dependency-injection_CN.md">中文</a>
</p>

# Gone Framework Dependency Injection Core Path Explanation

- [Gone Framework Dependency Injection Core Path Explanation](#gone-framework-dependency-injection-core-path-explanation)
	- [1. Component Definition and Goner Interface](#1-component-definition-and-goner-interface)
	- [2. Component Loading Process](#2-component-loading-process)
		- [2.1 Core Structure](#21-core-structure)
		- [2.2 Component Loading Flow](#22-component-loading-flow)
	- [3. Dependency Checking and Circular Dependency Detection](#3-dependency-checking-and-circular-dependency-detection)
		- [3.1 Dependency Collection](#31-dependency-collection)
		- [3.2 Component Installation](#32-component-installation)
	- [4. Field Injection Implementation and Provider Mechanism](#4-field-injection-implementation-and-provider-mechanism)
		- [4.1 Field Injection](#41-field-injection)
		- [4.2 Provider Mechanism](#42-provider-mechanism)
	- [5. Lifecycle Management](#5-lifecycle-management)
		- [5.1 Component Initialization](#51-component-initialization)
		- [5.2 Application Lifecycle](#52-application-lifecycle)
		- [5.3 Lifecycle Hooks](#53-lifecycle-hooks)
	- [6. Function Parameter Injection](#6-function-parameter-injection)
	- [7. Dependency Injection Process Summary](#7-dependency-injection-process-summary)


Gone is a lightweight Go language dependency injection framework that helps developers build modular, testable applications through a concise API and flexible component management mechanism. This article will provide a detailed explanation of the core path of dependency injection in the Gone framework, covering the complete process from component definition to dependency injection.

## 1. Component Definition and Goner Interface

All components in the Gone framework must implement the `Goner` interface, which is a marker interface used to identify components that can be managed by the Gone framework.

```go
// Goner is the basic interface that all components managed by Gone must implement
// It serves as a marker interface to identify types that can be loaded into the Gone container
type Goner interface {
	goneFlag()
}
```

To simplify component definition, Gone provides a `Flag` structure. Any type that embeds this structure automatically implements the `Goner` interface:

```go
// Flag is a marker structure used to identify components that can be managed by the gone framework
// Embedding this structure in other structures indicates that it can be used for gone's dependency injection
type Flag struct{}

func (g *Flag) goneFlag() {}
```

Component definition example:

```go
// Define a simple component
type MyComponent struct {
    gone.Flag  // Embed Flag to implement the Goner interface
    // Component fields
    Dependency *AnotherComponent `gone:"*"` // Declare dependencies using gone tags
}
```

## 2. Component Loading Process

The Gone framework loads components into the container through the `Core.Load` method. This process includes component registration, Provider detection, and dependency relationship establishment.

### 2.1 Core Structure

`Core` is the heart of the Gone framework, responsible for component loading, dependency injection, and lifecycle management:

```go
type Core struct {
	Flag
	coffins []*coffin

	nameMap            map[string]*coffin
	typeProviderMap    map[reflect.Type]*wrapProvider
	typeProviderDepMap map[reflect.Type]*coffin
	loaderMap          map[LoaderKey]bool
	log                Logger `gone:"*"`
}
```

### 2.2 Component Loading Flow

The `Core.Load` method is the entry point for component loading and accomplishes the following tasks:

1. Create a coffin wrapper for the component
2. Handle named component registration
3. Detect and register Providers
4. Apply loading options (such as default implementations, loading order, etc.)

```go
func (s *Core) Load(goner Goner, options ...Option) error {
	if goner == nil {
		return NewInnerError("goner cannot be nil - must provide a valid Goner instance", LoadedError)
	}
	co := newCoffin(goner)

	if namedGoner, ok := goner.(NamedGoner); ok {
		co.name = namedGoner.GonerName()
	}

	// Apply loading options
	for _, option := range options {
		if err := option.Apply(co); err != nil {
			return ToError(err)
		}
	}

	// Handle named component registration
	if co.name != "" {
		// Check for name conflicts and handle them
		// ...
	}

	// Add to component list
	s.coffins = append(s.coffins, co)

	// Detect and register Provider
	provider := tryWrapGonerToProvider(goner)
	if provider != nil {
		co.needInitBeforeUse = true
		co.provider = provider

		// Register Provider
		// ...
	}
	return nil
}
```

## 3. Dependency Checking and Circular Dependency Detection

Before initializing components, the Gone framework first checks dependency relationships to ensure there are no circular dependencies and to determine the optimal initialization order.

### 3.1 Dependency Collection

The `Core.Check` method is responsible for collecting all dependencies between components and detecting circular dependencies:

```go
func (s *Core) Check() ([]dependency, error) {
	// Collect dependency relationships
	depsMap, err := s.collectDeps()
	if err != nil {
		return nil, ToError(err)
	}

	// Check for circular dependencies and get the best initialization order
	deps, orders := checkCircularDepsAndGetBestInitOrder(depsMap)
	if len(deps) > 0 {
		return nil, circularDepsError(deps)
	}

	// Add field filling operations
	for _, co := range s.coffins {
		orders = append(orders, dependency{co, fillAction})
	}
	return RemoveRepeat(orders), nil
}
```

### 3.2 Component Installation

The `Core.Install` method executes component initialization in the determined order:

```go
func (s *Core) Install() error {
	// Get initialization order
	orders, err := s.Check()
	if err != nil {
		return ToError(err)
	}

	// Execute field filling and initialization in order
	for i, dep := range orders {
		if dep.action == fillAction {
			if err := s.safeFillOne(dep.coffin); err != nil {
				s.log.Debugf("failed to %s at order[%d]: %s", dep, i, err)
				return ToError(err)
			}
		}
		if dep.action == initAction {
			if err = s.safeInitOne(dep.coffin); err != nil {
				s.log.Debugf("failed to %s at order[%d]: %s", dep, i, err)
				return ToError(err)
			}
		}
	}
	return nil
}
```

## 4. Field Injection Implementation and Provider Mechanism

The Gone framework implements field injection through reflection mechanisms and supports various Provider mechanisms to create and provide dependency instances.

### 4.1 Field Injection

The `Core.fillOne` method is responsible for injecting dependencies into components:

```go
func (s *Core) fillOne(coffin *coffin) error {
	goner := coffin.goner

	// Call BeforeInit hook
	if initiator, ok := goner.(BeforeInitiatorNoError); ok {
		initiator.BeforeInit()
	}

	// Use reflection to get struct fields
	elem := reflect.TypeOf(goner).Elem()
	elemV := reflect.ValueOf(goner).Elem()

	// Iterate through all fields
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		v := elemV.Field(i)

		// Look for gone tag
		if tag, ok := field.Tag.Lookup(goneTag); ok {
			goneName, extend := ParseGoneTag(tag)
			if goneName == "" {
				goneName = defaultProviderName
			}

			// Get dependency
			co, err := s.getDepByName(goneName)
			if err != nil {
				return ToErrorWithMsg(err, fmt.Sprintf("failed to find dependency %q for field %q in type %q", goneName, field.Name, GetTypeName(elem)))
			}

			// Try to directly inject compatible type
			if IsCompatible(field.Type, co.goner) {
				v.Set(reflect.ValueOf(co.goner))
				continue
			}

			// Try using Provider to provide dependency
			if co.provider != nil && field.Type == co.provider.Type() {
				provide, err := co.provider.Provide(extend)
				if err != nil {
					return ToErrorWithMsg(err, fmt.Sprintf("provider %T failed to provide value for field %q in type %q", co.goner, field.Name, GetTypeName(elem)))
				} else if provide != nil {
					v.Set(reflect.ValueOf(provide))
					continue
				}
			}

			// Try using NamedProvider
			if provider, ok := co.goner.(NamedProvider); ok {
				// ...
			}

			// Try using StructFieldInjector
			if injector, ok := co.goner.(StructFieldInjector); ok {
				// ...
			}
		}
	}

	coffin.isFill = true
	return nil
}
```

### 4.2 Provider Mechanism

The Gone framework supports various Provider interfaces for creating and providing dependency instances:

```go
// Provider is a generic interface for providing dependencies of type T
type Provider[T any] interface {
	Goner
	Provide(tagConf string) (T, error)
}

// NamedProvider interface for creating dependencies based on name and type
type NamedProvider interface {
	NamedGoner
	Provide(tagConf string, t reflect.Type) (any, error)
}

// NoneParamProvider is a simplified Provider interface
type NoneParamProvider[T any] interface {
	Goner
	Provide() (T, error)
}
```

The `Core.Provide` method implements the logic for finding and creating dependencies:

```go
func (s *Core) Provide(tagConf string, t reflect.Type) (any, error) {
	// Try using type Provider
	if provider, ok := s.typeProviderMap[t]; ok && provider != nil {
		provide, err := provider.Provide(tagConf)
		if err != nil {
			s.log.Warnf("provider %T failed to provide value for type %s: %v",
				provider, GetTypeName(t), err)
		} else if provide != nil {
			return provide, nil
		}
	}

	// Try to find default implementation
	c := s.getDefaultCoffinByType(t)
	if c != nil {
		return c.goner, nil
	}

	// Handle special case for slice types
	if t.Kind() == reflect.Slice {
		// ...
	}

	return nil, NewInnerError(
		fmt.Sprintf("no provider or compatible type found for %s", GetTypeName(t)),
		NotSupport)
}
```

## 5. Lifecycle Management

The Gone framework provides complete component lifecycle management, including initialization, startup, and shutdown phases.

### 5.1 Component Initialization

The `Core.initOne` method is responsible for initializing a single component:

```go
func (s *Core) initOne(c *coffin) error {
	goner := c.goner
	if initiator, ok := goner.(InitiatorNoError); ok {
		initiator.Init()
	}
	if initiator, ok := goner.(Initiator); ok {
		if err := initiator.Init(); err != nil {
			return ToError(err)
		}
	}
	c.isInit = true
	return nil
}
```

### 5.2 Application Lifecycle

The `Application` structure is responsible for managing the entire application lifecycle, including startup and shutdown:

```go
type Application struct {
	Flag

	loader  *Core    `gone:"*"`
	daemons []Daemon `gone:"*"`

	beforeStartHooks []Process
	afterStartHooks  []Process
	beforeStopHooks  []Process
	afterStopHooks   []Process

	signal chan os.Signal
}
```

`Application` provides a series of methods to manage the application lifecycle:

```go
// Start the application
func (s *Application) start() {
	// Execute pre-start hooks
	for _, fn := range s.beforeStartHooks {
		fn()
	}

	// Start all daemons
	for _, daemon := range s.daemons {
		err := daemon.Start()
		if err != nil {
			panic(err)
		}
	}

	// Execute post-start hooks
	for _, fn := range s.afterStartHooks {
		fn()
	}
}

// Stop the application
func (s *Application) stop() {
	// Execute pre-stop hooks
	for _, fn := range s.beforeStopHooks {
		fn()
	}

	// Stop all daemons in reverse order
	for i := len(s.daemons) - 1; i >= 0; i-- {
		err := s.daemons[i].Stop()
		if err != nil {
			panic(err)
		}
	}

	// Execute post-stop hooks
	for _, fn := range s.afterStopHooks {
		fn()
	}
}
```

### 5.3 Lifecycle Hooks

The Gone framework provides various lifecycle hooks that allow components to execute custom logic at different stages of the application:

```go
// Define hook function type
type Process func()

// Pre-start hook
type BeforeStart func(Process)

// Post-start hook
type AfterStart func(Process)

// Pre-stop hook
type BeforeStop func(Process)

// Post-stop hook
type AfterStop func(Process)
```

Hook registration and execution order:

```go
// Register pre-start hook
func (s *Application) beforeStart(fn Process) {
	// Note: Pre-start hooks are executed in last-in-first-out order
	s.beforeStartHooks = append([]Process{fn}, s.beforeStartHooks...)
}

// Register post-start hook
func (s *Application) afterStart(fn Process) {
	// Note: Post-start hooks are executed in first-in-first-out order
	s.afterStartHooks = append(s.afterStartHooks, fn)
}
```

## 6. Function Parameter Injection

The Gone framework not only supports struct field injection but also function parameter injection, which allows direct use of dependencies in functions:

```go
// Function parameter injection
func (s *Core) InjectFuncParameters(fn any, injectBefore FuncInjectHook, injectAfter FuncInjectHook) (args []reflect.Value, err error) {
	ft := reflect.TypeOf(fn)

	if ft.Kind() != reflect.Func {
		return nil, NewInnerError(fmt.Sprintf("cannot inject parameters: expected a function, got %v", ft.Kind()), NotSupport)
	}

	in := ft.NumIn()

	for i := 0; i < in; i++ {
		pt := ft.In(i)
		paramName := fmt.Sprintf("parameter #%d (%s)", i+1, GetTypeName(pt))

		injected := false

		// Try using custom injection hook
		if injectBefore != nil {
			v := injectBefore(pt, i, false)
			if v != nil {
				args = append(args, reflect.ValueOf(v))
				injected = true
			}
		}

		// Try using standard dependency injection
		if !injected {
			if v, err := s.Provide("", pt); err != nil && !IsError(err, NotSupport) {
				return nil, ToErrorWithMsg(err, fmt.Sprintf("failed to inject %s in %s", paramName, GetFuncName(fn)))
			} else if v != nil {
				args = append(args, reflect.ValueOf(v))
				injected = true
			}
		}

		// Try to create and fill struct parameters
		if !injected {
			// ...
		}

		// Try using post-injection hook
		if injectAfter != nil {
			// ...
		}

		if !injected {
			return nil, NewInnerError(fmt.Sprintf("no suitable injector found for %s in %s", paramName, GetFuncName(fn)), NotSupport)
		}
	}
	return
}
```

Function wrapper for executing functions with injected parameters:

```go
func (s *Core) InjectWrapFunc(fn any, injectBefore FuncInjectHook, injectAfter FuncInjectHook) (func() []any, error) {
	args, err := s.InjectFuncParameters(fn, injectBefore, injectAfter)
	if err != nil {
		return nil, err
	}

	return func() (results []any) {
		values := reflect.ValueOf(fn).Call(args)
		for _, arg := range values {
			// Process return values
			// ...
			results = append(results, arg.Interface())
		}
		return results
	}, nil
}
```

## 7. Dependency Injection Process Summary

The core path of dependency injection in the Gone framework can be summarized in the following steps:

1. **Component Definition**: Implement the `Goner` interface by embedding the `gone.Flag` structure to make components manageable by the Gone framework.

2. **Component Loading**: Use the `Core.Load` or `Application.Load` method to load components into the container, with options to specify name, default implementation, etc.

3. **Dependency Checking**: The framework checks dependencies between components, ensures there are no circular dependencies, and determines the best initialization order.

4. **Dependency Injection**: The framework uses reflection to inject dependencies into each component, supporting various injection methods:
    - Direct injection of compatible types
    - Using Provider to create dependency instances
    - Using NamedProvider to create dependencies based on name and type
    - Using StructFieldInjector for custom injection logic

5. **Component Initialization**: Initialize components in the determined order, calling `BeforeInit` and `Init` methods.

6. **Application Startup**: Execute pre-start hooks, start all daemons, execute post-start hooks.

7. **Application Running**: The application runs normally, handling business logic.

8. **Application Shutdown**: Execute pre-stop hooks, stop all daemons in reverse order, execute post-stop hooks.

Through this series of steps, the Gone framework implements a flexible and powerful dependency injection mechanism, helping developers build modular and testable applications.