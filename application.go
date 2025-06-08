package gone

import (
	"os"
	"os/signal"
	"syscall"
)

// Application represents the core container and orchestrator for the Gone framework.
// Think of it as the "command center" or "central dispatch" of your application - it's like
// the conductor of an orchestra who knows every musician (component), their instruments (capabilities),
// and how they should work together to create beautiful music (your application).
//
// The Application acts as the central hub that:
// - Loads and manages all components (like a "personnel manager")
// - Handles dependency injection between components (like a "matchmaker")
// - Manages application lifecycle with hooks (like a "stage director")
// - Provides access to components via the GonerKeeper interface (like a "directory service")
//
// Design Philosophy:
// - Single Responsibility: Each Application instance manages one complete application context
// - Composition over Inheritance: Built by composing various specialized components
// - Encapsulation: Internal complexity is hidden behind simple, intuitive interfaces
//
// Key responsibilities:
// - Component registration and loading ("hiring and onboarding")
// - Dependency resolution and injection ("team building and collaboration")
// - Lifecycle management with hooks ("project management")
// - Graceful shutdown handling ("orderly dismissal")
type Application struct {
	Flag

	loader  *core    `gone:"*"`
	daemons []Daemon `gone:"*"`

	beforeStartHooks []Process
	afterStartHooks  []Process
	beforeStopHooks  []Process
	afterStopHooks   []Process

	signal chan os.Signal
}

// NewApp creates and initializes a new Application instance.
// Think of it as "founding a new company" where LoadFuncs are like "department setup plans"
// that define how to establish different parts of your application. Each LoadFunc knows
// how to "hire" and "organize" specific types of components.
//
// The Creation Process:
// - Establishes the "company headquarters" (Application instance)
// - Registers "department setup plans" (LoadFuncs)
// - Prepares the "organizational structure" for component management
//
// It creates an empty Application struct and calls init() to:
// 1. Initialize signal channel
// 2. Create new Core
// 3. Load core components like providers and default configure
// Returns the initialized Application instance ready for use.
func NewApp(loads ...LoadFunc) *Application {
	preparer := Application{}

	preparer.init()
	return preparer.Loads(loads...)
}

// Preparer is a type alias for Application, representing the main entry point for application setup and execution.
type Preparer = Application

// Prepare is alias for NewApp
func Prepare(loads ...LoadFunc) *Application {
	return NewApp(loads...)
}

func (s *Application) init() *Application {
	s.signal = make(chan os.Signal, 1)
	s.loader = newCore()

	s.
		Load(s, IsDefault()).
		Load(&BeforeStartProvider{}).
		Load(&AfterStartProvider{}).
		Load(&BeforeStopProvider{}).
		Load(&AfterStopProvider{})
	return s
}

// Load loads a Goner into the Application's loader with optional configuration options.
// Think of it as "hiring a new employee" where you bring a specific person (component)
// into your company (application) and give them their "employee handbook" (options).
// It wraps the Core.Load() method and panics if loading fails.
//
// The Hiring Process:
// - Verify the candidate has proper "credentials" (implements Goner interface)
// - Assign them a unique "employee ID" (internal tracking)
// - Set up their "workspace" and "job description" (configuration)
// - Add them to the "company directory" (component registry)
//
// Parameters:
//   - goner: The Goner instance to load - the "new hire"
//   - options: Optional configuration options for the Goner - the "employment terms"
//
// Available Options:
//   - Name(name string): Set custom name for the Goner
//   - IsDefault(): Mark this Goner as the default implementation
//   - OnlyForName(): Only register by name, not as provider
//   - ForceReplace(): Replace existing Goner with same name/type
//   - Order(order int): Set initialization order (lower runs first)
//   - FillWhenInit(): Fill dependencies during initialization
//
// Returns the Application instance for method chaining
func (s *Application) Load(goner Goner, options ...Option) *Application {
	err := s.loader.Load(goner, options...)
	if err != nil {
		panic(err)
	}
	return s
}

// Loads executes multiple LoadFuncs in sequence to load goner for Application
// Think of it as "batch hiring" where you bring multiple new employees into your
// company at once, like during a "recruitment drive" or "team expansion".
//
// The Batch Hiring Process:
// - Process each LoadFunc in sequence
// - Each LoadFunc acts like a "department setup plan"
// - Stop the process if any "hiring" fails
//
// Parameters:
//   - loads: Variadic LoadFunc parameters that will be executed in order - the "batch of hiring plans"
//
// Each LoadFunc typically loads goner components.
// If any LoadFunc fails during execution, it will trigger a panic.
//
// Returns:
//   - *Application: Returns the Application instance itself for method chaining
func (s *Application) Loads(loads ...LoadFunc) *Application {
	for _, fn := range loads {
		s.loader.MustLoadX(fn)
	}
	return s
}

// BeforeStart registers a function to be called before starting the application.
// Think of it as scheduling a "pre-opening meeting" where you can perform final preparations
// before your "business" officially opens its doors. The function will be executed before
// any daemons are started.
//
// Typical Use Cases:
// - Final system checks and validations
// - Cache warming and data preloading
// - External service connections
// - Configuration validation
//
// Returns the Application instance for method chaining.
func (s *Application) BeforeStart(fn Process) *Application {
	s.beforeStart(fn)
	return s
}

func (s *Application) beforeStart(fn Process) {
	s.beforeStartHooks = append([]Process{fn}, s.beforeStartHooks...)
}

// AfterStart registers a function to be called after starting the application.
// Think of it as scheduling a "grand opening celebration" or "post-launch activities"
// that happen after your "business" is officially open and running. The function will
// be executed after all daemons have been started.
//
// Typical Use Cases:
// - Success notifications and logging
// - Health check registrations
// - Monitoring and metrics setup
// - External service announcements
//
// Returns the Application instance for method chaining.
func (s *Application) AfterStart(fn Process) *Application {
	s.afterStart(fn)
	return s
}

func (s *Application) afterStart(fn Process) {
	s.afterStartHooks = append(s.afterStartHooks, fn)
}

// BeforeStop registers a function to be called before stopping the application.
// Think of it as scheduling "closing preparations" where you perform necessary tasks
// before your "business" officially closes. The function will be executed before
// any daemons are stopped.
//
// Typical Use Cases:
// - Graceful connection closures
// - Data persistence and state saving
// - Resource cleanup and release
// - Shutdown notifications
//
// Returns the Application instance for method chaining.
func (s *Application) BeforeStop(fn Process) *Application {
	s.beforeStop(fn)
	return s
}

func (s *Application) beforeStop(fn Process) {
	s.beforeStopHooks = append([]Process{fn}, s.beforeStopHooks...)
}

// AfterStop registers a function to be called after stopping the application.
// Think of it as "post-closure activities" that happen after your "business" has
// officially closed its doors. The function will be executed after all daemons
// have been stopped.
//
// Typical Use Cases:
// - Final cleanup and resource release
// - Shutdown success logging
// - External service deregistration
// - Final state persistence
//
// Returns the Application instance for method chaining.
func (s *Application) AfterStop(fn Process) *Application {
	s.afterStop(fn)
	return s
}

func (s *Application) afterStop(fn Process) {
	s.afterStopHooks = append(s.afterStopHooks, fn)
}

// WaitEnd blocks until the application receives a termination signal (SIGINT, SIGTERM, or SIGQUIT).
// Think of it as a "security guard" who watches the door and waits for the "closing time signal".
// This method listens for termination signals (like Ctrl+C or system shutdown) and returns when
// one is received, allowing for graceful shutdown procedures.
//
// Signal Monitoring:
// - SIGINT: Usually triggered by Ctrl+C ("manual closing")
// - SIGTERM: System shutdown or process termination ("scheduled closing")
// - SIGQUIT: Quit signal ("emergency closing")
//
// Returns the Application instance for method chaining.
func (s *Application) WaitEnd() *Application {
	signal.Notify(s.signal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-s.signal
	return s
}

// End triggers application termination by sending a SIGINT signal.
// Think of it as the "official closing procedure" where you send the "closing signal"
// to initiate graceful shutdown. This method triggers application termination by
// sending a SIGINT signal.
//
// Returns the Application instance for method chaining.
func (s *Application) End() *Application {
	s.signal <- syscall.SIGINT
	return s
}

func (s *Application) start() {
	for _, fn := range s.beforeStartHooks {
		fn()
	}

	for _, daemon := range s.daemons {
		err := daemon.Start()
		if err != nil {
			panic(err)
		}
	}

	for _, fn := range s.afterStartHooks {
		fn()
	}
}

func (s *Application) stop() {
	for _, fn := range s.beforeStopHooks {
		fn()
	}

	for i := len(s.daemons) - 1; i >= 0; i-- {
		err := s.daemons[i].Stop()
		if err != nil {
			panic(err)
		}
	}

	for _, fn := range s.afterStopHooks {
		fn()
	}
}

func (s *Application) install() {
	err := s.loader.Install()
	if err != nil {
		panic(err)
	}
}

func (s *Application) collectHooks() {
	coffins := s.loader.iKeeper.getAllCoffins()
	for _, co := range coffins {
		if co.goner != nil {
			if start, ok := co.goner.(BeforeStarter); ok {
				s.beforeStart(func() {
					start.BeforeStart()
				})
			}
			if afterStart, ok := co.goner.(AfterStarter); ok {
				s.afterStart(func() {
					afterStart.AfterStart()
				})
			}
			if stop, ok := co.goner.(BeforeStopper); ok {
				s.beforeStop(func() {
					stop.BeforeStop()
				})
			}
			if afterStop, ok := co.goner.(AfterStopper); ok {
				s.afterStop(func() {
					afterStop.AfterStop()
				})
			}
		}
	}
}

// Run initializes the application, injects dependencies into the provided function,
// executes it, and then performs cleanup.
// Think of it as "opening your business for a specific task" - you unlock the doors,
// turn on all systems, perform the specific work, then properly close everything down.
// The function can have dependencies that will be automatically injected.
// Panics if dependency injection or execution fails.
//
// The Complete Business Day Process:
// 1. "System setup" - Install and initialize all components
// 2. "Team coordination" - Collect and register lifecycle hooks
// 3. "Open for business" - Execute start procedures
// 4. "Main work" - Execute provided functions with dependency injection
// 5. "Proper closure" - Execute stop procedures
//
// Parameters:
//   - funcList: The function to execute with injected dependencies - the "main business tasks"
func (s *Application) Run(funcList ...any) {
	s.install()
	s.collectHooks()
	s.start()

	var options []RunOption
	for _, fn := range funcList {
		if r, ok := fn.(RunOption); ok {
			options = append(options, r)
			continue
		}

		f, err := s.loader.InjectWrapFunc(fn, nil, nil)
		if err != nil {
			panic(err)
		}
		_ = f()
	}

	for _, o := range options {
		o.Apply(s)
	}

	s.stop()
}

// Serve initializes the application, starts all daemons, and waits for termination signal.
// Think of it as "opening your business for continuous operation" - you unlock the doors,
// start all services, and keep the business running until you receive a "closing signal".
// After receiving termination signal, performs cleanup by stopping all daemons.
//
// The Continuous Operation Process:
// - Same as Run() but includes OpWaitEnd() to wait for termination signals
// - Ideal for long-running applications like web servers or background services
//
func (s *Application) Serve(funcList ...any) {
	funcList = append(funcList, OpWaitEnd())
	s.Run(funcList...)
}

type RunOption interface {
	Apply(*Application)
}

type waitEnd struct{}

func (waitEnd) Apply(s *Application) {
	s.WaitEnd()
}

func OpWaitEnd() RunOption {
	return waitEnd{}
}

type TestFlag interface {
	forTest()
}

type testFlag struct {
	Flag
}

func (*testFlag) forTest() {}

// Test runs the application in test mode with dependency injection.
// Think of it as setting up a "testing laboratory" where you can experiment with your
// components in a controlled environment. This method is designed for testing scenarios
// where you need to execute test functions with proper dependency injection.
//
// The Testing Laboratory Process:
// - Loads a test flag to indicate test mode
// - Executes the test function using Run() with full dependency injection
//
// Key Features:
// - Full dependency injection support
// - Simplified execution for testing purposes
// - Test mode indication via testFlag
//
func (s *Application) Test(fn any) {
	s.Load(&testFlag{})
	s.Run(fn)
}

// RunTest Deprecated, use Test instead
func RunTest(fn any, priests ...LoadFunc) {
	NewApp(priests...).Test(fn)
}
