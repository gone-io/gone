package gone

import (
	"os"
	"os/signal"
	"syscall"
)

// Prepare creates and initializes a new Preparer instance.
// It creates an empty Preparer struct and calls init() to:
// 1. Initialize signal channel
// 2. Create new Core
// 3. Load core components like providers and default configure
// Returns the initialized Preparer instance ready for use.
func Prepare(loads ...LoadFunc) *Preparer {
	preparer := Preparer{}

	preparer.init()
	for _, fn := range loads {
		err := fn(preparer.loader)
		if err != nil {
			panic(err)
		}
	}
	return &preparer
}

var Default = Prepare()

type Preparer struct {
	Flag

	loader  *Core    `gone:"*"`
	daemons []Daemon `gone:"*"`

	beforeStartHooks []Process
	afterStartHooks  []Process
	beforeStopHooks  []Process
	afterStopHooks   []Process

	signal chan os.Signal
}

func (s *Preparer) init() *Preparer {
	s.signal = make(chan os.Signal, 1)
	s.loader = NewCore()

	s.
		Load(s, IsDefault()).
		Load(&BeforeStartProvider{}).
		Load(&AfterStartProvider{}).
		Load(&BeforeStopProvider{}).
		Load(&AfterStopProvider{}).
		Load(&ConfigProvider{}).
		Load(&EnvConfigure{}, Name("configure"), IsDefault(new(Configure)), OnlyForName()).
		Load(defaultLog, IsDefault(new(Logger)))
	return s
}

// Load loads a Goner into the Preparer's loader with optional configuration options.
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
// Returns the Preparer instance for method chaining
func (s *Preparer) Load(goner Goner, options ...Option) *Preparer {
	err := s.loader.Load(goner, options...)
	if err != nil {
		panic(err)
	}
	return s
}

func Load(goner Goner, options ...Option) *Preparer {
	return Default.Load(goner, options...)
}

func (s *Preparer) Loads(loads ...LoadFunc) *Preparer {
	for _, fn := range loads {
		err := fn(s.loader)
		if err != nil {
			panic(err)
		}
	}
	return s
}

func Loads(loads ...LoadFunc) *Preparer {
	return Default.Loads(loads...)
}

// BeforeStart registers a function to be called before starting the application.
// The function will be executed before any daemons are started.
// Returns the Preparer instance for method chaining.
func (s *Preparer) BeforeStart(fn Process) *Preparer {
	s.beforeStart(fn)
	return s
}

func (s *Preparer) beforeStart(fn Process) {
	s.beforeStartHooks = append(s.beforeStartHooks, fn)
}

// AfterStart registers a function to be called after starting the application.
// The function will be executed after all daemons have been started.
// Returns the Preparer instance for method chaining.
func (s *Preparer) AfterStart(fn Process) *Preparer {
	s.afterStart(fn)
	return s
}

func (s *Preparer) afterStart(fn Process) {
	s.afterStartHooks = append(s.afterStartHooks, fn)
}

// BeforeStop registers a function to be called before stopping the application.
// The function will be executed before any daemons are stopped.
// Returns the Preparer instance for method chaining.
func (s *Preparer) BeforeStop(fn Process) *Preparer {
	s.beforeStop(fn)
	return s
}

func (s *Preparer) beforeStop(fn Process) {
	s.beforeStopHooks = append(s.beforeStopHooks, fn)
}

// AfterStop registers a function to be called after stopping the application.
// The function will be executed after all daemons have been stopped.
// Returns the Preparer instance for method chaining.
func (s *Preparer) AfterStop(fn Process) *Preparer {
	s.afterStop(fn)
	return s
}

func (s *Preparer) afterStop(fn Process) {
	s.afterStopHooks = append(s.afterStopHooks, fn)
}

// WaitEnd blocks until the application receives a termination signal (SIGINT, SIGTERM, or SIGQUIT).
// Returns the Preparer instance for method chaining.
func (s *Preparer) WaitEnd() *Preparer {
	signal.Notify(s.signal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-s.signal
	return s
}

// End triggers application termination by sending a SIGINT signal.
// Returns the Preparer instance for method chaining.
func (s *Preparer) End() *Preparer {
	s.signal <- syscall.SIGINT
	return s
}

func (s *Preparer) start() {
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

func (s *Preparer) stop() {
	for _, fn := range s.beforeStopHooks {
		fn()
	}

	for _, daemon := range s.daemons {
		err := daemon.Stop()
		if err != nil {
			panic(err)
		}
	}

	for _, fn := range s.afterStopHooks {
		fn()
	}
}

func (s *Preparer) install() {
	err := s.loader.Install()
	if err != nil {
		panic(err)
	}
}

// Run initializes the application, injects dependencies into the provided function,
// executes it, and then performs cleanup.
// The function can have dependencies that will be automatically injected.
// Panics if dependency injection or execution fails.
//
// Parameters:
//   - fn: The function to execute with injected dependencies
func (s *Preparer) Run(fn any) {
	s.install()
	s.start()

	f, err := s.loader.InjectWrapFunc(fn, nil, nil)
	if err != nil {
		panic(err)
	}
	_ = f()
	s.stop()
}

func Run(fn any) {
	Default.Run(fn)
}

// Serve initializes the application, starts all daemons, and waits for termination signal.
// After receiving termination signal, performs cleanup by stopping all daemons.
func (s *Preparer) Serve() {
	s.install()
	s.start()
	s.WaitEnd()
	s.stop()
}

func Serve() {
	Default.Serve()
}

type TestFlag interface {
	forTest()
}

type testFlag struct {
	Flag
}

func (*testFlag) forTest() {}

func (s *Preparer) Test(fn any) {
	s.Load(&testFlag{})
	s.Run(fn)
}

// Test for run tests
func Test(fn any) {
	Default.Test(fn)
}

// RunTest Deprecated, use Test instead
func RunTest(fn any, priests ...LoadFunc) {
	Prepare(priests...).Test(fn)
}
