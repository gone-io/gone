package gone

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

// BeforeStarter interface defines components that need to perform actions before the application starts.
// Components implementing this interface will have their BeforeStart() method called during Gone's startup phase,
// before any daemons are started but after all components have been initialized.
//
// The BeforeStart() method should:
// - Perform any setup needed before daemons start
// - Initialize resources that daemons depend on
// - Handle errors internally since no error is returned
//
// Example usage:
//
//	type MyComponent struct {
//	    gone.Flag
//	    logger *Logger `gone:"*"`
//	}
//
//	func (c *MyComponent) BeforeStart() {
//	    c.logger.Info("Preparing for application start")
//	    // perform pre-start setup...
//	}
type BeforeStarter interface {
	BeforeStart()
}

// AfterStarter interface defines components that need to perform actions after the application starts.
// Components implementing this interface will have their AfterStart() method called during Gone's startup phase,
// after all daemons have been started successfully.
//
// The AfterStart() method should:
// - Perform any setup that requires all services to be running
// - Register with external services or health checks
// - Handle errors internally since no error is returned
//
// Example usage:
//
//	type MyComponent struct {
//	    gone.Flag
//	    healthCheck *HealthCheck `gone:"*"`
//	}
//
//	func (c *MyComponent) AfterStart() {
//	    c.healthCheck.Register()
//	    // perform post-start setup...
//	}
type AfterStarter interface {
	AfterStart()
}

// BeforeStopper interface defines components that need to perform actions before the application stops.
// Components implementing this interface will have their BeforeStop() method called during Gone's shutdown phase,
// before any daemons are stopped but after termination signal is received.
//
// The BeforeStop() method should:
// - Perform cleanup tasks while all services are still running
// - Save important state or data
// - Gracefully disconnect from external services
// - Handle errors internally since no error is returned
//
// Example usage:
//
//	type MyComponent struct {
//	    gone.Flag
//	    cache *Cache `gone:"*"`
//	}
//
//	func (c *MyComponent) BeforeStop() {
//	    c.cache.Flush()
//	    // perform pre-stop cleanup...
//	}
type BeforeStopper interface {
	BeforeStop()
}

// AfterStopper interface defines components that need to perform actions after the application stops.
// Components implementing this interface will have their AfterStop() method called during Gone's shutdown phase,
// after all daemons have been stopped successfully.
//
// The AfterStop() method should:
// - Perform final cleanup tasks
// - Release any remaining resources
// - Log shutdown completion
// - Handle errors internally since no error is returned
//
// Example usage:
//
//	type MyComponent struct {
//	    gone.Flag
//	    logger *Logger `gone:"*"`
//	}
//
//	func (c *MyComponent) AfterStop() {
//	    c.logger.Info("Application shutdown complete")
//	    // perform final cleanup...
//	}
type AfterStopper interface {
	AfterStop()
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

// Hook functions allow components to properly initialize, cleanup, and coordinate
// with other components during the application lifecycle.

// Process represents a function that performs some operation without taking parameters or returning values.
// It is commonly used for Hook functions in the application lifecycle, such as BeforeStart, AfterStart,
// BeforeStop and AfterStop Hook.
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

// BeforeStart is a HookReg function type that can be injected into Goners to register callbacks
// that will execute before the application starts.
//
// Example usage:
// ```go
//
//	type XGoner struct {
//	    Flag
//	    before BeforeStart `gone:"*"` // Inject the BeforeStart HookReg
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

// AfterStart is a HookReg function type that can be injected into Goners to register callbacks
// that will execute after the application starts.
//
// Example usage:
// ```go
//
//	type XGoner struct {
//	    Flag
//	    after AfterStart `gone:"*"` // Inject the AfterStart HookReg
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

// BeforeStop is a HookReg function type that can be injected into Goners to register callbacks
// that will execute before the application stops.
//
// Example usage:
// ```go
//
//	type XGoner struct {
//	    Flag
//	    before BeforeStop `gone:"*"` // Inject the BeforeStop HookReg
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

// AfterStop is a HookReg function type that can be injected into Goners to register callbacks
// that will execute after the application stops.
//
// Example usage:
// ```go
//
//	type XGoner struct {
//	    Flag
//	    after AfterStop `gone:"*"` // Inject the AfterStop HookReg
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
