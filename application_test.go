package gone_test

import (
	"errors"
	"go.uber.org/mock/gomock"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gone-io/gone/v2"
)

// MockDaemon implements Daemon interface for testing
type MockDaemon struct {
	gone.Flag
	startCalled bool
	stopCalled  bool
	startError  error
	stopError   error
	mu          sync.Mutex
}

func (m *MockDaemon) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startCalled = true
	return m.startError
}

func (m *MockDaemon) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopCalled = true
	return m.stopError
}

func TestPreparer_Lifecycle(t *testing.T) {
	var hooksCalled []string
	var mu sync.Mutex

	addHookCall := func(name string) {
		mu.Lock()
		defer mu.Unlock()
		hooksCalled = append(hooksCalled, name)
	}

	preparer := gone.NewApp()

	// Register hooks
	preparer.BeforeStart(func() {
		addHookCall("beforeStart")
	})
	preparer.AfterStart(func() {
		addHookCall("afterStart")
	})
	preparer.BeforeStop(func() {
		addHookCall("beforeStop")
	})
	preparer.AfterStop(func() {
		addHookCall("afterStop")
	})

	// Add a test daemon
	daemon := &MockDaemon{}
	preparer.Load(daemon)

	// Run in a goroutine and end after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		preparer.End()
	}()

	preparer.Serve()

	// Verify hook execution order
	expectedOrder := []string{
		"beforeStart",
		"afterStart",
		"beforeStop",
		"afterStop",
	}

	if !reflect.DeepEqual(expectedOrder, hooksCalled) {
		t.Errorf("Hooks called in wrong order. Expected %v, got %v", expectedOrder, hooksCalled)
	}

	// Verify daemon methods were called
	if !daemon.startCalled {
		t.Error("Daemon Start() was not called")
	}
	if !daemon.stopCalled {
		t.Error("Daemon Stop() was not called")
	}
}

func TestPreparer_DaemonErrors(t *testing.T) {
	tests := []struct {
		name      string
		daemon    *MockDaemon
		wantPanic bool
	}{
		{
			name: "Start error",
			daemon: &MockDaemon{
				startError: errors.New("start error"),
			},
			wantPanic: true,
		},
		{
			name: "Stop error",
			daemon: &MockDaemon{
				stopError: errors.New("stop error"),
			},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preparer := gone.NewApp()
			preparer.Load(tt.daemon)

			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic but got none")
					}
				}()

				go func() {
					time.Sleep(100 * time.Millisecond)
					preparer.End()
				}()
				preparer.Serve()
			}
		})
	}
}

func TestPreparer_SignalHandling(t *testing.T) {
	preparer := gone.NewApp()

	// Test SIGINT
	go func() {
		time.Sleep(100 * time.Millisecond)
		preparer.End()
	}()

	start := time.Now()
	preparer.Serve()
	duration := time.Since(start)

	if duration >= 200*time.Millisecond {
		t.Errorf("Signal handling took too long: %v", duration)
	}
}

func TestPreparer_MultipleHooks(t *testing.T) {
	preparer := gone.NewApp()
	var counter int
	var mu sync.Mutex

	increment := func() {
		mu.Lock()
		defer mu.Unlock()
		counter++
	}

	// Register multiple hooks for each phase
	for i := 0; i < 3; i++ {
		preparer.BeforeStart(increment)
		preparer.AfterStart(increment)
		preparer.BeforeStop(increment)
		preparer.AfterStop(increment)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		preparer.End()
	}()

	preparer.Serve()

	if counter != 12 {
		t.Errorf("Not all hooks were called, expected 12, got %d", counter)
	}
}

func TestPreparer_LoadErrors(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*gone.Application)
		wantPanic bool
	}{
		{
			name: "Duplicate named component",
			setup: func(p *gone.Application) {
				p.Load(&Worker{name: "test"})
				p.Load(&Worker{name: "test"})
			},
			wantPanic: true,
		},
		{
			name: "Valid components",
			setup: func(p *gone.Application) {
				p.Load(&Worker{name: "worker1"})
				p.Load(&Worker{name: "worker2"})
			},
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preparer := gone.NewApp()
			didPanic := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						didPanic = true
					}
				}()
				tt.setup(preparer)
			}()
			if tt.wantPanic != didPanic {
				t.Errorf("Test %s: wantPanic = %v, got panic = %v", tt.name, tt.wantPanic, didPanic)
			}
		})
	}
}

func TestPreparer_RunWithDependencies(t *testing.T) {
	preparer := gone.NewApp()

	worker1 := &Worker{name: "worker1"}
	worker2 := &Worker{name: "worker2"}
	boss := &Boss{name: "boss"}

	preparer.Load(worker1).
		Load(worker2).
		Load(boss)

	var executed bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()
		preparer.Run(func(b *Boss) {
			if b.name != "boss" {
				t.Errorf("Expected boss name to be 'boss', got %s", b.name)
			}
			if b.first == nil {
				t.Error("Expected first worker to not be nil")
			}
			if b.second == nil {
				t.Error("Expected second worker to not be nil")
			}
			if len(b.workers) != 2 {
				t.Errorf("Expected 2 workers, got %d", len(b.workers))
			}
			executed = true
		})
	}()

	if !executed {
		t.Error("Run function was not executed")
	}
}

func TestPreparer_DefaultInstance(t *testing.T) {
	if gone.Default == nil {
		t.Error("Default preparer instance should not be nil")
	}

	worker := &Worker{name: "test"}
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()
		gone.Default.Load(worker)
	}()
}

func TestPreparer_Loads(t *testing.T) {
	preparer := gone.NewApp()

	// Test successful loads
	loadFn1 := func(core gone.Loader) error {
		return core.Load(&Worker{name: "worker1"})
	}
	loadFn2 := func(core gone.Loader) error {
		return core.Load(&Worker{name: "worker2"})
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()
		preparer.Loads(loadFn1, loadFn2)
	}()

	// Test load function that returns error
	errorLoadFn := func(core gone.Loader) error {
		return errors.New("load error")
	}

	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		preparer.Loads(errorLoadFn)
	}()

	if !didPanic {
		t.Error("Expected Loads to panic with error load function")
	}
}

func TestPreparer_Test(t *testing.T) {
	var testFuncCalled bool

	testFunc := func(flag gone.TestFlag) {
		if flag == nil {
			t.Error("TestFlag should not be nil")
		}
		testFuncCalled = true
	}

	preparer := gone.NewApp()
	preparer.Test(testFunc)

	if !testFuncCalled {
		t.Error("Test function was not called")
	}
}

func TestPreparer_GlobalFunctions(t *testing.T) {
	// Test global Load function
	worker := &Worker{name: "global-worker"}
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()
		gone.Load(worker)
	}()

	// Test global Loads function
	loadFn := func(core gone.Loader) error {
		return core.Load(&Worker{name: "global-worker2"})
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()
		gone.Loads(loadFn)
	}()

	// Test global Run function
	var runCalled bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()
		gone.Run(func() {
			runCalled = true
		})
	}()
	if !runCalled {
		t.Error("Global Run function did not execute")
	}

	// Test global Test function
	var testCalled bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()
		gone.Test(func(flag gone.TestFlag) {
			if flag == nil {
				t.Error("TestFlag should not be nil")
			}
			testCalled = true
		})
	}()
	if !testCalled {
		t.Error("Global Test function did not execute")
	}
}

func TestPreparer_RunTest(t *testing.T) {
	var testFuncCalled bool

	testFunc := func(flag gone.TestFlag) {
		if flag == nil {
			t.Error("TestFlag should not be nil")
		}
		testFuncCalled = true
	}

	loadFn := func(core gone.Loader) error {
		return core.Load(&Worker{name: "test-worker"})
	}

	gone.RunTest(testFunc, loadFn)

	if !testFuncCalled {
		t.Error("RunTest function was not called")
	}
}

func TestPreparer_PrepareWithLoads(t *testing.T) {
	loadFn1 := func(core gone.Loader) error {
		return core.Load(&Worker{name: "worker1"})
	}
	loadFn2 := func(core gone.Loader) error {
		return core.Load(&Worker{name: "worker2"})
	}

	var preparer *gone.Application
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()
		preparer = gone.NewApp(loadFn1, loadFn2)
	}()

	if preparer == nil {
		t.Error("Application should not be nil")
	}

	// Test prepare with error load function
	errorLoadFn := func(core gone.Loader) error {
		return errors.New("load error")
	}

	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		gone.NewApp(errorLoadFn)
	}()

	if !didPanic {
		t.Error("Expected NewApp to panic with error load function")
	}
}

func TestPreparer_ServeGlobal(t *testing.T) {
	// Test global Serve function
	go func() {
		time.Sleep(100 * time.Millisecond)
		gone.Default.End()
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()
		gone.Serve()
	}()
}

func TestApplication_Loads(t *testing.T) {
	loadFn := func(core gone.Loader) error {
		return core.Load(&Worker{name: "worker"})
	}

	gone.
		NewApp(loadFn).
		Loads(loadFn, loadFn).
		Run(func(works []*Worker) {
			if len(works) != 1 {
				t.Errorf("Expected 3 workers, got %d", 1)
			}
		})
}

func TestLoadsUsage(t *testing.T) {
	loadFn := func(loader gone.Loader) error {
		loader.
			MustLoad(&Worker{name: "worker"}).
			MustLoad(&Worker{name: "worker2"}).
			MustLoad(&Worker{name: "worker3"}, gone.Name("boss"))
		return nil
	}
	gone.
		NewApp(loadFn).
		Loads(loadFn, loadFn).
		Run(func(works []*Worker, i struct {
			boss *Worker `gone:"boss"`
		}) {
			if len(works) != 3 {
				t.Errorf("Expected 3 workers, got %d", 3)
			}
			if i.boss.name != "worker3" {
				t.Errorf("Expected boss name to be worker3, got %s", i.boss.name)
			}
		})
}

func TestMustLoadPanic(t *testing.T) {
	loadFn := func(loader gone.Loader) error {
		loader.
			MustLoad(&Worker{name: "worker"}).
			MustLoad(&Worker{name: "worker2"}).
			MustLoad(&Worker{name: "worker3"}, gone.Name("boss")).
			MustLoad(&Worker{name: "worker4"}, gone.Name("boss"))
		return nil
	}
	err := gone.SafeExecute(func() error {
		gone.NewApp(loadFn).Run()
		return nil
	})
	if err == nil {
		t.Errorf("Expected panic, got nil")
		return
	}
	if !strings.Contains(err.Error(), "goner with name \"boss\" is already loaded") {
		t.Errorf("Expected duplicate name error, got %s", err.Error())
	}
}

func TestPrepare(t *testing.T) {
	gone.Prepare().Run()
}

func TestEnd(t *testing.T) {
	gone.
		Default.
		AfterStart(func() {
			time.Sleep(1 * time.Millisecond)
			gone.End()
		}).
		Serve()
}

func TestApplication_collectHooks(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	var order, beforeStartOrder, afterStartOrder, beforeStopOrder, afterStopOrder int

	starter := gone.NewMockBeforeStarter(controller)
	starter.EXPECT().BeforeStart().Do(func() {
		order++
		beforeStartOrder = order
	})
	afterStarter := gone.NewMockAfterStarter(controller)
	afterStarter.EXPECT().AfterStart().Do(func() {
		order++
		afterStartOrder = order
	})

	stoper := gone.NewMockBeforeStoper(controller)
	stoper.EXPECT().BeforeStop().Do(func() {
		order++
		beforeStopOrder = order
	})

	afterStoper := gone.NewMockAfterStoper(controller)

	afterStoper.EXPECT().AfterStop().Do(func() {
		order++
		afterStopOrder = order
	})
	gone.
		NewApp().
		Load(stoper).
		Load(afterStoper).
		Load(starter).
		Load(afterStarter).
		Run(func() {
			if beforeStartOrder != 1 {
				t.Errorf("beforeStartOrder = %d, want %d", beforeStartOrder, 1)
			}
			if afterStartOrder != 2 {
				t.Errorf("afterStartOrder = %d, want %d", afterStartOrder, 2)
			}
			if beforeStopOrder != 0 {
				t.Errorf("beforeStopOrder = %d, want %d", beforeStopOrder, 0)
			}
			if afterStopOrder != 0 {
				t.Errorf("afterStopOrder = %d, want %d", afterStopOrder, 0)
			}
		})

	if beforeStopOrder != 3 {
		t.Errorf("beforeStopOrder = %d, want %d", beforeStopOrder, 3)
	}
	if afterStopOrder != 4 {
		t.Errorf("afterStopOrder = %d, want %d", afterStopOrder, 4)
	}
}
