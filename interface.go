package gone

import (
	"reflect"
	"sync"
)

// Goner is the base interface that all components managed by Gone must implement.
// It acts as a "component identity card" - any component wanting to be managed by the Gone framework
// must hold this "identity card". This is a clever design using a private method to ensure
// only "official channels" can obtain this "identity card".
//
// Any struct that embeds the Flag struct automatically implements this interface.
// This allows Gone to verify that components are properly configured for dependency injection.
// Just like getting an ID card requires going to the designated office, becoming a Goner
// can only be achieved by embedding the `gone.Flag` "official seal".
//
// Design Benefits:
//   - Security: Prevents "counterfeit" components from entering the system
//   - Consistency: Ensures all components follow the same "registration" process
//   - Control: Framework has complete control over component lifecycle
//
// Example usage:
//
//	type MyComponent struct {
//	    gone.Flag  // Embeds Flag to implement Goner - the "identity card"
//	}
type Goner interface {
	goneFlag()
}

// Component is an alias for Goner.
type Component = Goner

// NamedGoner extends the Goner interface to add naming capability to components.
// Components implementing this interface can be registered and looked up by name in the Gone container,
// like having a "business card" with a specific name that others can use to find them.
//
// The GonerName() method should return a unique string identifier for the component.
// This name acts as the component's "business card" and can be used when:
// - Loading the component with explicit name
// - Looking up dependencies by name using `gone:"name"` tags
// - Registering multiple implementations of the same interface
// - Distinguishing between different instances of the same type
//
// Example usage:
//
//	type MyNamedComponent struct {
//	    gone.Flag
//	}
//
//	func (c *MyNamedComponent) GonerName() string {
//	    return "myComponent"  // This component's "business card"
//	}
type NamedGoner interface {
	Goner
	GonerName() string
}

// Provider is a generic interface for components that can provide dependencies of type T.
// Think of it as a "smart factory" that can create specific types of components on demand.
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
// The Provider acts like a "specialized craftsman" who knows how to create specific types
// of components based on the requirements (tagConf) provided.
//
// Parameters for Provide:
//   - tagConf: Configuration string from the struct tag that requested this dependency,
//             like a "work order" specifying what kind of component is needed
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
//	    return &Config{}, nil  // Creates a new Config based on requirements
//	}
type Provider[T any] interface {
	Goner
	Provide(tagConf string) (T, error)
}

// NoneParamProvider is a simplified Provider interface for components that provide dependencies without requiring tag configuration.
// Think of it as a "simple factory" that creates standard products without needing special instructions.
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
// This provider acts like a "standard craftsman" who creates the same type of component
// every time without needing special instructions or customization.
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
// Think of it as a "master craftsman" with a business card who can create any type of component
// based on specific requirements. It extends NamedGoner to support named component registration
// and provides a more flexible Provide method that can create dependencies of any type.
//
// The interface requires:
//   - Embedding the NamedGoner interface to support named component registration
//   - Implementing Provide() to create dependencies based on tag config and requested type
//
// This provider acts like a "versatile craftsman" who can adapt their skills to create
// different types of components based on the specific type requested and configuration provided.
//
// Parameters for Provide:
//   - tagConf: Configuration string from the struct tag that requested this dependency,
//             like a "detailed work order" specifying requirements
//   - t: The `reflect.Type` of the dependency being requested,
//        like a "blueprint" showing what type of component to create
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
//	    // Create and return instance based on the type blueprint
//	    return &Config{}, nil
//	}
type NamedProvider interface {
	NamedGoner
	Provide(tagConf string, t reflect.Type) (any, error)
}

// StructFieldInjector is an interface for components that can inject dependencies into struct fields.
// Think of it as a "specialized installer" who knows how to install specific components into
// struct fields based on installation instructions. It extends NamedGoner to support named
// component registration and provides a method to inject dependencies into struct fields
// based on tag configuration and field information.
//
// The interface requires:
//   - Embedding the NamedGoner interface to support named component registration
//   - Implementing Inject() to inject dependencies into struct fields
//
// This injector acts like a "field specialist" who can precisely install components
// into specific struct fields based on the field's requirements and configuration.
//
// Parameters for Inject:
//   - tagConf: Configuration string from the struct tag that requested this dependency,
//             like "installation instructions" specifying how to inject
//   - field: The `reflect.StructField` that requires injection,
//           like the "installation location" specification
//   - fieldValue: The actual field value to be modified during injection
//
// Returns:
//   - error: Any error that occurred during injection
type StructFieldInjector interface {
	NamedGoner
	Inject(tagConf string, field reflect.StructField, fieldValue reflect.Value) error
}

// Daemon represents a long-running service component that can be started and stopped.
// Think of it as a "background service worker" that runs continuously to provide specific
// functionality, like a web server, database connection pool, or message queue processor.
//
// Lifecycle Management:
// Daemons are started in order of registration when Application.Serve() or Application.start() is called.
// When the application receives a termination signal, daemons are stopped in reverse order
// to ensure proper cleanup and resource management.
//
// Error Handling:
// - If Start() returns an error, the application will panic to prevent inconsistent state
// - If Stop() returns an error, the application will panic to ensure proper cleanup
//
// Example usage:
// ```go
//
//	type MyDaemon struct {
//	    gone.Flag
//	}
//
//	func (d *MyDaemon) Start() error {
//	    // Initialize and start the daemon's main functionality
//	    // Like starting a web server or opening database connections
//	    return nil
//	}
//
//	func (d *MyDaemon) Stop() error {
//	    // Gracefully shut down the daemon and clean up resources
//	    // Like closing connections and saving state
//	    return nil
//	}
//
// ```
type Daemon interface {
	Start() error
	Stop() error
}

// FuncInjector provides methods for injecting dependencies into function parameters.
// Think of it as an "intelligent assistant" that automatically identifies what parameters
// a function needs, then finds the corresponding components from the "warehouse" and
// automatically "feeds" them to the function. It's like ordering takeout where the
// delivery person automatically brings all the dishes according to your order.
//
// The interface requires implementing:
//   - InjectFuncParameters: Injects dependencies into function parameters
//   - InjectWrapFunc: Wraps a function with dependency injection
//
// Magic Principles:
//   - Parameter Recognition: Automatically analyzes function signatures to understand required parameter types
//   - Smart Matching: Finds matching instances from registered components
//   - Auto Invocation: Calls the target function with found components as parameters
//
// Parameters for both methods:
//   - fn: The function to inject dependencies into
//   - injectBefore: Optional hook called before standard injection
//   - injectAfter: Optional hook called after standard injection
//
// Typical Applications:
//   - Controller Methods: Automatic parameter injection for web request handlers
//   - Utility Functions: Tool function calls requiring multiple dependencies
//   - Testing Scenarios: Automatic dependency assembly for test functions
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
// Think of it as an experienced "assembly worker" who can automatically identify which fields
// in a struct need "component installation", then find suitable components from the "parts warehouse"
// and precisely "install" them in the specified locations. It's like assembling a computer where
// the worker automatically installs corresponding hardware based on the motherboard's interfaces.
//
// The interface requires implementing:
//   - InjectStruct: Injects dependencies into struct fields based on gone tags
//
// Work Flow:
//   - Field Scanning: Checks struct fields with specific tags
//   - Component Matching: Finds corresponding components based on field types and tag information
//   - Precise Installation: "Installs" found components into corresponding fields
//
// Application Scenarios:
//   - Component Initialization: Automatically assembles dependencies after component creation
//   - Test Preparation: Automatically injects mock dependencies for test objects
//   - Dynamic Assembly: Supplements dependencies for existing objects at runtime
//
// Example usage:
// ```go
//
//	type MyGoner struct {
//	    gone.Flag
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
	//   - goner: Should be a struct pointer.
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
// Think of it as a "component registration number" that uses an internal counter to ensure
// each loaded component gets a unique identifier, like a social security number for components.
//
// The LoaderKey is used to:
// - Track which components have been loaded (like a "registration database")
// - Prevent duplicate loading of components (avoid "double registration")
// - Provide a way to check component load status ("registration verification")
//
// This ensures the framework can efficiently manage component lifecycle and prevent conflicts.
type LoaderKey struct{ id uint64 }

// LoadFunc represents a function that can load components into a Gone container.
// Think of it as a "professional moving company" that knows how to properly "pack" and
// "relocate" specific types of components into the Gone framework. Each LoadFunc knows
// how to correctly handle the loading process for particular component types.
//
// LoadFunc Work Flow:
//   - Package Components: Prepare business components for loading
//   - Transport and Load: Load components into the framework through Loader
//   - Confirm Checklist: Return loading results (success or error)
//
// It takes a Loader interface as parameter to allow loading additional dependencies,
// enabling LoadFuncs to form a tree-like structure for complex component hierarchies.
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
// Think of it as a company's "HR onboarding specialist" responsible for handling
// "onboarding procedures" for new components. It assigns each new component a "employee ID",
// establishes their "personnel file", and officially includes them in the Gone framework's
// "employee roster".
//
// Onboarding Process:
//   - Identity Verification: Confirms component implements Goner interface (holds "ID card")
//   - Assign ID: Allocates unique identifier for the component
//   - Establish Records: Records component type, dependencies, and other information
//   - Archive Management: Stores component information in the framework's management system
//
// Usage Scenarios:
//   - Application Startup: Batch loading of core application components
//   - Plugin Loading: Dynamic loading of plugin components
//   - Test Environment: Loading specific mock components for testing
//
// The interface requires implementing:
//   - Load: Loads a component into the container with optional configuration
//   - MustLoad: Loads a component and panics on error (for critical components)
//   - MustLoadX: Flexible loading that handles both components and LoadFuncs
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
// Think of it as a super-intelligent "archive manager" who knows the "home address" of every
// component in the system. Whether you want to find someone by "name" or by "profession" (type),
// it can quickly locate them. It provides dynamic access to components at runtime, allowing
// components to be looked up by either name or type.
//
// The interface requires implementing:
//   - GetGonerByName: Retrieves a component by its registered name
//   - GetGonerByType: Retrieves a component by its type
//   - GetGonerByPattern: Retrieves components matching a pattern
//
// Practical Scenarios:
//   - Dynamic Lookup: Find specific components at runtime as needed
//   - Precise Location: Quickly locate target components by name or type
//   - Pattern Matching: Use wildcard patterns to batch retrieve components
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
//	    // Get component by name - like looking up someone's "business card"
//	    if svc := m.keeper.GetGonerByName("service"); svc != nil {
//	        // Use the service
//	    }
//
//	    // Get component by type - like finding someone by "profession"
//	    if logger := m.keeper.GetGonerByType(reflect.TypeOf(&Logger{})); logger != nil {
//	        // Use the logger
//	    }
//	    return nil
//	}
//
// ```
type GonerKeeper interface {
	// GetGonerByName retrieves a component by its name.
	// Like looking up someone's "business card" in a directory, this method finds components
	// by their registered name. The name should match either the component's explicit name
	// set via gone.Name() option or the name returned by its GonerName() method if it implements NamedGoner.
	//
	// Parameters:
	//   - name: The name of the component to retrieve (the "business card" identifier)
	//
	// Returns:
	//   - any: The component instance if found, nil otherwise
	GetGonerByName(name string) any

	// GetGonerByType retrieves a component by its type.
	// Like finding someone by their "profession" in a company directory, this method
	// locates components by their type. The type should match either the exact type
	// of the component or an interface type that the component implements.
	//
	// Parameters:
	//   - t: The reflect.Type of the component to retrieve (the "profession" specification)
	//
	// Returns:
	//   - any: The component instance if found, nil otherwise
	GetGonerByType(t reflect.Type) any

	// GetGonerByPattern retrieves components matching a pattern.
	// Like searching for "all engineers whose names start with 'John'", this method
	// finds multiple components that match both a type and a name pattern.
	//
	// Parameters:
	//   - t: The reflect.Type of components to search for
	//   - pattern: The name pattern to match (supports wildcards like * and ?)
	//
	// Returns:
	//   - []any: Array of component instances that match the criteria
	GetGonerByPattern(t reflect.Type, pattern string) []any
}

type Keeper = GonerKeeper
