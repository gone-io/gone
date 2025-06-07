package gone

import (
	"os"
	"os/signal"
	"syscall"
)

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
// It wraps the Core.Load() method and panics if loading fails.
//
// Parameters:
//   - goner: The Goner instance to load
//   - options: Optional configuration options for the Goner
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
// Parameters:
//   - loads: Variadic LoadFunc parameters that will be executed in order
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
// The function will be executed before any daemons are started.
// Returns the Application instance for method chaining.
func (s *Application) BeforeStart(fn Process) *Application {
	s.beforeStart(fn)
	return s
}

func (s *Application) beforeStart(fn Process) {
	s.beforeStartHooks = append([]Process{fn}, s.beforeStartHooks...)
}

// AfterStart registers a function to be called after starting the application.
// The function will be executed after all daemons have been started.
// Returns the Application instance for method chaining.
func (s *Application) AfterStart(fn Process) *Application {
	s.afterStart(fn)
	return s
}

func (s *Application) afterStart(fn Process) {
	s.afterStartHooks = append(s.afterStartHooks, fn)
}

// BeforeStop registers a function to be called before stopping the application.
// The function will be executed before any daemons are stopped.
// Returns the Application instance for method chaining.
func (s *Application) BeforeStop(fn Process) *Application {
	s.beforeStop(fn)
	return s
}

func (s *Application) beforeStop(fn Process) {
	s.beforeStopHooks = append([]Process{fn}, s.beforeStopHooks...)
}

// AfterStop registers a function to be called after stopping the application.
// The function will be executed after all daemons have been stopped.
// Returns the Application instance for method chaining.
func (s *Application) AfterStop(fn Process) *Application {
	s.afterStop(fn)
	return s
}

func (s *Application) afterStop(fn Process) {
	s.afterStopHooks = append(s.afterStopHooks, fn)
}

// WaitEnd blocks until the application receives a termination signal (SIGINT, SIGTERM, or SIGQUIT).
// Returns the Application instance for method chaining.
func (s *Application) WaitEnd() *Application {
	signal.Notify(s.signal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-s.signal
	return s
}

// End triggers application termination by sending a SIGINT signal.
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
			if stop, ok := co.goner.(BeforeStoper); ok {
				s.beforeStop(func() {
					stop.BeforeStop()
				})
			}
			if afterStop, ok := co.goner.(AfterStoper); ok {
				s.afterStop(func() {
					afterStop.AfterStop()
				})
			}
		}
	}
}

// Run initializes the application, injects dependencies into the provided function,
// executes it, and then performs cleanup.
// The function can have dependencies that will be automatically injected.
// Panics if dependency injection or execution fails.
//
// Parameters:
//   - funcList: The function to execute with injected dependencies
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
// After receiving termination signal, performs cleanup by stopping all daemons.
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

func (s *Application) Test(fn any) {
	s.Load(&testFlag{})
	s.Run(fn)
}

// RunTest Deprecated, use Test instead
func RunTest(fn any, priests ...LoadFunc) {
	NewApp(priests...).Test(fn)
}
