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

// Component is an alias for Goner.
type Component = Goner

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
//	func (c *MyNamedComponent) GonerName() string {
//	    return "myComponent"
//	}
type NamedGoner interface {
	Goner
	GonerName() string
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
//	    preparer *Application
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
//   - t: The `reflect.Type` of the dependency being requested
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

// StructFieldInjector is an interface for components that can inject dependencies into struct fields.
// It extends NamedGoner to support named component registration and provides a method to inject dependencies
// into struct fields based on tag configuration and field information.
//
// The interface requires:
//   - Embedding the NamedGoner interface to support named component registration
//   - Implementing Inject() to inject dependencies into struct fields
//
// Parameters for Inject:
//   - tagConf: Configuration string from the struct tag that requested this dependency
//   - field: The `reflect.StructField` that requires injection
type StructFieldInjector interface {
	NamedGoner
	Inject(tagConf string, field reflect.StructField, fieldValue reflect.Value) error
}

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
// Daemons are started in order of registration when Application.Serve() or Application.start() is called.
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

// StructInjector defines the interface for components that can inject dependencies into struct fields.
// It provides a single method to perform dependency injection on struct instances that implement the Goner interface.
//
// The interface requires implementing:
//   - InjectStruct: Injects dependencies into struct fields based on gone tags
//
// Example usage:
// ```go
//
//	type MyGoner struct {
//	    Flag
//	    Service *MyService `gone:"*"`  // Field to be injected
//	}
//
//	injector := &Core{}
//	goner := &MyGoner{}
//	err := injector.InjectStruct(goner)
//	if err != nil {
//	    panic(err)
//	}
//	// MyGoner.Service is now injected
//
// ```
//
// The InjectStruct method analyzes the struct's fields, looking for `gone` tags,
// and injects the appropriate dependencies based on the tag configuration.
type StructInjector interface {
	// InjectStruct performs dependency injection on the provided Goner struct.
	// It scans the struct's fields for `gone` tags and injects the appropriate dependencies.
	//
	// Parameters:
	//   - goner: The struct instance to inject dependencies into. Must implement Goner interface.
	//
	// Returns:
	//   - error: Any error that occurred during injection
	InjectStruct(goner any) error
}

var (
	keyMtx     sync.Mutex
	keyCounter uint64
)

// LoaderKey is a unique identifier for tracking loaded components in the Gone container.
// It uses an internal counter to ensure each loaded component gets a unique key.
//
// The LoaderKey is used to:
// - Track which components have been loaded
// - Prevent duplicate loading of components
// - Provide a way to check component load status
type LoaderKey struct{ id uint64 }

// LoadFunc represents a function that can load components into a Gone container.
// It takes a Loader interface as parameter to allow loading additional dependencies.
//
// Example usage:
// ```go
//
//	func loadComponents(l Loader) error {
//	    if err := l.Load(&ServiceA{}); err != nil {
//	        return err
//	    }
//	    if err := l.Load(&ServiceB{}); err != nil {
//	        return err
//	    }
//	    return nil
//	}
//
// ```
type LoadFunc = func(Loader) error

type MustLoadFunc = func(Loader)

// Loader defines the interface for loading components into the Gone container.
// It provides methods to load new components and check if components are already loaded.
//
// The interface requires implementing:
//   - Load: Loads a component into the container with optional configuration
//   - Loaded: Checks if a component is already loaded
type Loader interface {
	// Load adds a component to the Gone container with optional configuration.
	//
	// Parameters:
	//   - goner: The component to load. Must implement Goner interface.
	//   - options: Optional configuration for how the component should be loaded.
	//
	// Returns:
	//   - error: Any error that occurred during loading
	Load(goner Goner, options ...Option) error

	MustLoadX(x any) Loader

	// MustLoad adds a component to the Gone container with optional configuration.
	// If an error occurs during loading, it panics.
	//
	// Parameters:
	//   - goner: The component to load. Must implement Goner interface.
	//   - options: Optional configuration for how the component should be loaded.
	//
	// Returns:
	//   - Loader: The Loader instance for further loading operations
	MustLoad(goner Goner, options ...Option) Loader

	// Loaded checks if a component identified by the given LoaderKey has been loaded.
	//
	// Parameters:
	//   - LoaderKey: The unique identifier for the component to check.
	//
	// Returns:
	//   - bool: true if the component is loaded, false otherwise
	Loaded(LoaderKey) bool
}

// GonerKeeper interface defines methods for retrieving components from the Gone container.
// It provides dynamic access to components at runtime, allowing components to be looked up
// by either name or type.
//
// The interface requires implementing:
//   - GetGonerByName: Retrieves a component by its registered name
//   - GetGonerByType: Retrieves a component by its type
//
// Example usage:
// ```go
//
//	type MyComponent struct {
//	    gone.Flag
//	    keeper gone.GonerKeeper `gone:"*"`
//	}
//
//	func (m *MyComponent) Init() error {
//	    // Get component by name
//	    if svc := m.keeper.GetGonerByName("service"); svc != nil {
//	        // Use the service
//	    }
//
//	    // Get component by type
//	    if logger := m.keeper.GetGonerByType(reflect.TypeOf(&Logger{})); logger != nil {
//	        // Use the logger
//	    }
//	    return nil
//	}
//
// ```
type GonerKeeper interface {
	// GetGonerByName retrieves a component by its name.
	// The name should match either the component's explicit name set via gone.Name() option
	// or the name returned by its GonerName() method if it implements NamedGoner.
	//
	// Parameters:
	//   - name: The name of the component to retrieve
	//
	// Returns:
	//   - any: The component instance if found, nil otherwise
	GetGonerByName(name string) any

	// GetGonerByType retrieves a component by its type.
	// The type should match either the exact type of the component or
	// an interface type that the component implements.
	//
	// Parameters:
	//   - t: The reflect.Type of the component to retrieve
	//
	// Returns:
	//   - any: The component instance if found, nil otherwise
	GetGonerByType(t reflect.Type) any

	GetGonerByPattern(t reflect.Type, pattern string) []any
}

type Keeper = GonerKeeper
