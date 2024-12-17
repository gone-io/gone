package gone

import (
	"reflect"
	"sync"
)

// Goner is the base interface that all components managed by Gone must implement.
// It acts as a marker interface to identify types that can be loaded into the Gone container.
//
// Any struct that embeds the Flag struct automatically implements this interface.
// This allows Gone to verify that components are properly configured for dependency injection.
//
// Example usage:
//
//	type MyComponent struct {
//	    gone.Flag  // Embeds Flag to implement Goner
//	}
type Goner interface {
	goneFlag()
}

// NamedGoner extends the Goner interface to add naming capability to components.
// Components implementing this interface can be registered and looked up by name in the Gone container.
//
// The Name() method should return a unique string identifier for the component.
// This name can be used when:
// - Loading the component with explicit name
// - Looking up dependencies by name using `gone:"name"` tags
// - Registering multiple implementations of the same interface
//
// Example usage:
//
//	type MyNamedComponent struct {
//	    gone.Flag
//	}
//
//	func (c *MyNamedComponent) Name() string {
//	    return "myComponent"
//	}
type NamedGoner interface {
	Goner
	Name() string
}

// Initiator interface defines components that need initialization after dependencies are injected.
// Components implementing this interface will have their Init() method called during Gone's initialization phase.
// Init() is called after all dependencies are filled and BeforeInit() hooks (if any) have completed.
//
// The Init() method should:
// - Perform any required setup or validation
// - Initialize internal state
// - Establish connections to external services
// - Return an error if initialization fails
//
// Example usage:
//
//	type MyComponent struct {
//	    gone.Flag
//	    db *Database `gone:"*"`
//	}
//
//	func (c *MyComponent) Init() error {
//	    return c.db.Connect()
//	}
type Initiator interface {
	Init() error
}

// InitiatorNoError interface defines components that need initialization but don't return errors.
// Similar to Initiator interface, but Init() does not return an error.
// Components implementing this interface will have their Init() method called during Gone's initialization phase,
// after dependencies are filled and BeforeInit() hooks (if any) have completed.
//
// The Init() method should:
// - Perform any required setup or validation
// - Initialize internal state
// - Establish connections to external services
// - Handle errors internally rather than returning them
//
// Example usage:
//
//	type MyComponent struct {
//	    gone.Flag
//	    logger *Logger `gone:"*"`
//	}
//
//	func (c *MyComponent) Init() {
//	    c.logger.Info("Initializing MyComponent")
//	    // perform initialization...
//	}
type InitiatorNoError interface {
	Init()
}

// BeforeInitiator interface defines components that need pre-initialization before regular initialization.
// Components implementing this interface will have their BeforeInit() method called during Gone's initialization phase,
// before dependencies are filled and before Init() is called.
//
// The BeforeInit() method should:
// - Perform any setup needed before dependencies are injected
// - Initialize basic internal state that doesn't depend on other components
// - Return an error if pre-initialization fails
//
// Example usage:
//
//	type MyComponent struct {
//	    gone.Flag
//	    config *Config
//	}
//
//	func (c *MyComponent) BeforeInit() error {
//	    // Setup basic state before dependencies are filled
//	    c.config = &Config{}
//	    return nil
//	}
type BeforeInitiator interface {
	BeforeInit() error
}

// BeforeInitiatorNoError interface defines components that need pre-initialization but don't return errors.
// Similar to BeforeInitiator interface, but BeforeInit() does not return an error.
// Components implementing this interface will have their BeforeInit() method called during Gone's initialization phase,
// before dependencies are filled and before Init() is called.
//
// The BeforeInit() method should:
// - Perform any setup needed before dependencies are injected
// - Initialize basic internal state that doesn't depend on other components
// - Handle errors internally rather than returning them
//
// Example usage:
//
//	type MyComponent struct {
//	    gone.Flag
//	    config *Config
//	}
//
//	func (c *MyComponent) BeforeInit() {
//	    // Setup basic state before dependencies are filled
//	    c.config = &Config{}
//	}
type BeforeInitiatorNoError interface {
	BeforeInit()
}

// Provider is a generic interface for components that can provide dependencies of type T.
// While not directly dependent on Gone's implementation, this interface helps developers
// write correct dependency providers that can be registered in the Gone container.
//
// Type Parameters:
//   - T: The type of dependency this provider creates
//
// The interface requires:
//   - Embedding the Goner interface to mark it as a Gone component
//   - Implementing Provide() to create and return instances of type T
//
// Parameters for Provide:
//   - tagConf: Configuration string from the struct tag that requested this dependency
//
// Returns:
//   - T: The created dependency instance
//   - error: Any error that occurred during creation
//
// Example usage:
//
//	type ConfigProvider struct {
//	    gone.Flag
//	}
//
//	func (p *ConfigProvider) Provide(tagConf string) (*Config, error) {
//	    return &Config{}, nil
//	}
type Provider[T any] interface {
	Goner
	Provide(tagConf string) (T, error)
}

// NoneParamProvider is a simplified Provider interface for components that provide dependencies without requiring tag configuration.
// Like Provider[T], this interface is not directly dependent on Gone's implementation but serves as a guide
// for writing simpler providers when tag configuration is not needed.
//
// Type Parameters:
//   - T: The type of dependency this provider creates
//
// The interface requires:
//   - Embedding the Goner interface to mark it as a Gone component
//   - Implementing Provide() to create and return instances of type T
//
// Returns:
//   - T: The created dependency instance
//   - error: Any error that occurred during creation
//
// Example usage:
//
//	type BeforeStartProvider struct {
//	    gone.Flag
//	    preparer *Preparer
//	}
//
//	func (p *BeforeStartProvider) Provide() (BeforeStart, error) {
//	    return p.preparer.beforeStart, nil
//	}
type NoneParamProvider[T any] interface {
	Goner
	Provide() (T, error)
}

// NamedProvider is an interface for providers that can create dependencies based on name and type.
// It extends NamedGoner to support named component registration and provides a more flexible Provide method
// that can create dependencies of any type.
//
// The interface requires:
//   - Embedding the NamedGoner interface to support named component registration
//   - Implementing Provide() to create dependencies based on tag config and requested type
//
// Parameters for Provide:
//   - tagConf: Configuration string from the struct tag that requested this dependency
//   - t: The reflect.Type of the dependency being requested
//
// Returns:
//   - any: The created dependency instance of the requested type
//   - error: Any error that occurred during creation
//
// Example usage:
//
//	type ConfigProvider struct {
//	    gone.Flag
//	}
//
//	func (p *ConfigProvider) Provide(tagConf string, t reflect.Type) (any, error) {
//	    // Create and return instance based on t
//	    return &Config{}, nil
//	}
type NamedProvider interface {
	NamedGoner
	Provide(tagConf string, t reflect.Type) (any, error)
}

// Gone Lifecycle:
//
// 1. Load: Components are loaded into the Gone container using the Load() method.
//    - Components are registered by name and/or type
//    - Provider components are registered to provide dependencies
//    - Configuration options like order and defaults are applied
//
// 2. Check: Gone validates the component configuration and determines initialization order
//    - Checks for circular dependencies between components that implement Init()
//    - Analyzes dependency graph to find optimal initialization sequence
//    - Validates all required dependencies exist
//    - Ensures no duplicate registrations
//    - Detects and prevents deadlocks in initialization order
//
// 3. Fill and Init: Dependencies are injected and components are initialized
//    - Components are filled with their dependencies in order
//    - BeforeInit() hooks are called if implemented
//    - Init() is called on components that implement Initiator
//    - Components are marked as initialized
//
// 4. Start: The application and its daemons are started
//    - BeforeStart hooks are executed
//    - Daemons are started in order based on Order() value
//    - AfterStart hooks are executed
//
// 5. End: The application runs until termination is triggered
//    - Waits for SIGINT, SIGTERM or SIGQUIT signal
//    - Can be triggered manually via End() method
//
// 6. Stop: Components are gracefully shut down
//    - BeforeStop hooks are executed
//    - Daemons are stopped in reverse order
//    - AfterStop hooks are executed
//    - Application terminates
//
// Hook Functions:
// Gone provides several hook functions that components can use to execute code at specific lifecycle points:
//
// - BeforeInit/BeforeInitNoError: Called before component initialization
//   Usage: Implement BeforeInitiator or BeforeInitiatorNoError interface
//
// - Init/InitNoError: Called during component initialization
//   Usage: Implement Initiator or InitiatorNoError interface
//
// - BeforeStart: Executed before application startup
//   Usage: Inject BeforeStart type and register callback functions
//
// - AfterStart: Executed after all components have started
//   Usage: Inject AfterStart type and register callback functions
//
// - BeforeStop: Executed before components begin shutting down
//   Usage: Inject BeforeStop type and register callback functions
//
// - AfterStop: Executed after all components have stopped
//   Usage: Inject AfterStop type and register callback functions
//
// Hook functions allow components to properly initialize, cleanup, and coordinate
// with other components during the application lifecycle.

// Process represents a function that performs some operation without taking parameters or returning values.
// It is commonly used for hook functions in the application lifecycle, such as BeforeStart, AfterStart,
// BeforeStop and AfterStop hooks.
//
// Example usage:
// ```go
//
//	type XGoner struct {
//	    Flag
//	    beforeStart BeforeStart `gone:"*"`
//	}
//
//	func (x *XGoner) Init() error {
//	    x.beforeStart(func() {
//	        // This is a Process function
//	        fmt.Println("Before application starts")
//	    })
//	    return nil
//	}
//
// ```
type Process func()

// BeforeStart is a hook function type that can be injected into Goners to register callbacks
// that will execute before the application starts.
//
// Example usage:
// ```go
//
//	type XGoner struct {
//	    Flag
//	    before BeforeStart `gone:"*"` // Inject the BeforeStart hook
//	}
//
//	func (x *XGoner) Init() error {
//	    // Register a callback to run before application start
//	    x.before(func() {
//	        fmt.Println("before start")
//	    })
//	    return nil
//	}
//
// ```
//
// The registered callbacks will be executed in registration order before any daemons are started.
// This allows components to perform initialization tasks that must complete before the application
// begins its main operations.
type BeforeStart func(Process)

// AfterStart is a hook function type that can be injected into Goners to register callbacks
// that will execute after the application starts.
//
// Example usage:
// ```go
//
//	type XGoner struct {
//	    Flag
//	    after AfterStart `gone:"*"` // Inject the AfterStart hook
//	}
//
//	func (x *XGoner) Init() error {
//	    // Register a callback to run after application start
//	    x.after(func() {
//	        fmt.Println("after start")
//	    })
//	    return nil
//	}
//
// ```
//
// The registered callbacks will be executed in registration order after all daemons have been started.
// This allows components to perform tasks that require all services to be running.
type AfterStart func(Process)

// BeforeStop is a hook function type that can be injected into Goners to register callbacks
// that will execute before the application stops.
//
// Example usage:
// ```go
//
//	type XGoner struct {
//	    Flag
//	    before BeforeStop `gone:"*"` // Inject the BeforeStop hook
//	}
//
//	func (x *XGoner) Init() error {
//	    // Register a callback to run before application stop
//	    x.before(func() {
//	        fmt.Println("before stop")
//	    })
//	    return nil
//	}
//
// ```
//
// The registered callbacks will be executed in registration order before any daemons are stopped.
// This allows components to perform cleanup tasks while services are still running.
type BeforeStop func(Process)

// AfterStop is a hook function type that can be injected into Goners to register callbacks
// that will execute after the application stops.
//
// Example usage:
// ```go
//
//	type XGoner struct {
//	    Flag
//	    after AfterStop `gone:"*"` // Inject the AfterStop hook
//	}
//
//	func (x *XGoner) Init() error {
//	    // Register a callback to run after application stop
//	    x.after(func() {
//	        fmt.Println("after stop")
//	    })
//	    return nil
//	}
//
// ```
//
// The registered callbacks will be executed in registration order after all daemons have been stopped.
// This allows components to perform final cleanup tasks after all services have been shut down.
type AfterStop func(Process)

// Daemon represents a long-running service component that can be started and stopped.
//
// Example usage:
// ```go
//
//	type MyDaemon struct {
//	    Flag
//	}
//
//	func (d *MyDaemon) Start() error {
//	    // Initialize and start the daemon
//	    return nil
//	}
//
//	func (d *MyDaemon) Stop() error {
//	    // Clean up and stop the daemon
//	    return nil
//	}
//
// ```
//
// Daemons are started in order of registration when Preparer.Serve() or Preparer.start() is called.
// The Start() method should initialize and start the daemon's main functionality.
// If Start() returns an error, the application will panic.
//
// When the application receives a termination signal, daemons are stopped in reverse order
// by calling their Stop() methods. The Stop() method should gracefully shut down the daemon
// and clean up any resources. If Stop() returns an error, the application will panic.
type Daemon interface {
	Start() error
	Stop() error
}

// FuncInjector provides methods for injecting dependencies into function parameters.
//
// The interface requires implementing:
//   - InjectFuncParameters: Injects dependencies into function parameters
//   - InjectWrapFunc: Wraps a function with dependency injection
//
// Parameters for both methods:
//   - fn: The function to inject dependencies into
//   - injectBefore: Optional hook called before standard injection
//   - injectAfter: Optional hook called after standard injection
//
// Example usage:
//
//	injector := &Core{}
//	fn := func(svc *MyService) error {
//	    return nil
//	}
//
//	wrapped, err := injector.InjectWrapFunc(fn, nil, nil)
//	if err != nil {
//	    panic(err)
//	}
//	results := wrapped()
type FuncInjector interface {
	// InjectFuncParameters injects dependencies into function parameters by:
	// 1. Using injectBefore hook if provided
	// 2. Using standard dependency injection
	// 3. Creating and filling struct parameters if needed
	// 4. Using injectAfter hook if provided
	// Returns the injected parameter values or error if injection fails
	InjectFuncParameters(fn any, injectBefore FuncInjectHook, injectAfter FuncInjectHook) (args []reflect.Value, err error)

	// InjectWrapFunc wraps a function with dependency injection.
	// It injects dependencies into the function parameters and returns a wrapper function that:
	// 1. Calls the original function with injected parameters
	// 2. Converts return values to []any, handling nil interface values appropriately
	// Returns wrapper function and error if injection fails
	InjectWrapFunc(fn any, injectBefore FuncInjectHook, injectAfter FuncInjectHook) (func() []any, error)
}

var (
	keyMtx     sync.Mutex
	keyCounter uint64
)

// LoaderKey is a throwaway value you can use as a key to a ContextManager
type LoaderKey struct{ id uint64 }

// GenLoaderKey will return a brand new, never-before-used LoaderKey
func GenLoaderKey() LoaderKey {
	keyMtx.Lock()
	defer keyMtx.Unlock()
	keyCounter += 1
	return LoaderKey{id: keyCounter}
}

type Loader interface {
	Load(goner Goner, options ...Option) error

	Loaded(LoaderKey) bool
}

type LoadFunc func(Loader) error

type GonerKeeper interface {
	GetGonerByName(name string) any
	GetGonerByType(t reflect.Type) any
}
