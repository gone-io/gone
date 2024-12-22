package gone_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
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

	preparer := gone.Prepare()

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

	assert.Equal(t, expectedOrder, hooksCalled, "Hooks called in wrong order")

	// Verify daemon methods were called
	assert.True(t, daemon.startCalled, "Daemon Start() was not called")
	assert.True(t, daemon.stopCalled, "Daemon Stop() was not called")
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
			preparer := gone.Prepare()
			preparer.Load(tt.daemon)

			if tt.wantPanic {
				assert.Panics(t, func() {
					go func() {
						time.Sleep(100 * time.Millisecond)
						preparer.End()
					}()
					preparer.Serve()
				})
			}
		})
	}
}

func TestPreparer_SignalHandling(t *testing.T) {
	preparer := gone.Prepare()

	// Test SIGINT
	go func() {
		time.Sleep(100 * time.Millisecond)
		preparer.End()
	}()

	start := time.Now()
	preparer.Serve()
	duration := time.Since(start)

	assert.Less(t, duration, 200*time.Millisecond, "Signal handling took too long")
}

func TestPreparer_MultipleHooks(t *testing.T) {
	preparer := gone.Prepare()
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

	assert.Equal(t, 12, counter, "Not all hooks were called")
}

func TestPreparer_LoadErrors(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*gone.Preparer)
		wantPanic bool
	}{
		{
			name: "Duplicate named component",
			setup: func(p *gone.Preparer) {
				p.Load(&Worker{name: "test"})
				p.Load(&Worker{name: "test"})
			},
			wantPanic: true,
		},
		{
			name: "Valid components",
			setup: func(p *gone.Preparer) {
				p.Load(&Worker{name: "worker1"})
				p.Load(&Worker{name: "worker2"})
			},
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preparer := gone.Prepare()
			if tt.wantPanic {
				assert.Panics(t, func() {
					tt.setup(preparer)
				})
			} else {
				assert.NotPanics(t, func() {
					tt.setup(preparer)
				})
			}
		})
	}
}

func TestPreparer_RunWithDependencies(t *testing.T) {
	preparer := gone.Prepare()

	worker1 := &Worker{name: "worker1"}
	worker2 := &Worker{name: "worker2"}
	boss := &Boss{name: "boss"}

	preparer.Load(worker1).
		Load(worker2).
		Load(boss)

	var executed bool
	assert.NotPanics(t, func() {
		preparer.Run(func(b *Boss) {
			assert.Equal(t, "boss", b.name)
			assert.NotNil(t, b.first)
			assert.NotNil(t, b.second)
			assert.Equal(t, 2, len(b.workers))
			executed = true
		})
	})

	assert.True(t, executed, "Run function was not executed")
}

func TestPreparer_DefaultInstance(t *testing.T) {
	assert.NotNil(t, gone.Default, "Default preparer instance should not be nil")

	// Test that Default instance is properly initialized
	worker := &Worker{name: "test"}
	assert.NotPanics(t, func() {
		gone.Default.Load(worker)
	})
}

func TestPreparer_Loads(t *testing.T) {
	preparer := gone.Prepare()

	// Test successful loads
	loadFn1 := func(core gone.Loader) error {
		return core.Load(&Worker{name: "worker1"})
	}
	loadFn2 := func(core gone.Loader) error {
		return core.Load(&Worker{name: "worker2"})
	}

	assert.NotPanics(t, func() {
		preparer.Loads(loadFn1, loadFn2)
	})

	// Test load function that returns error
	errorLoadFn := func(core gone.Loader) error {
		return errors.New("load error")
	}

	assert.Panics(t, func() {
		preparer.Loads(errorLoadFn)
	})
}

func TestPreparer_Test(t *testing.T) {
	var testFuncCalled bool

	testFunc := func(flag gone.TestFlag) {
		assert.NotNil(t, flag)
		testFuncCalled = true
	}

	preparer := gone.Prepare()
	preparer.Test(testFunc)

	assert.True(t, testFuncCalled, "Test function was not called")
}

func TestPreparer_GlobalFunctions(t *testing.T) {
	// Test global Load function
	worker := &Worker{name: "global-worker"}
	assert.NotPanics(t, func() {
		gone.Load(worker)
	})

	// Test global Loads function
	loadFn := func(core gone.Loader) error {
		return core.Load(&Worker{name: "global-worker2"})
	}
	assert.NotPanics(t, func() {
		gone.Loads(loadFn)
	})

	// Test global Run function
	var runCalled bool
	assert.NotPanics(t, func() {
		gone.Run(func() {
			runCalled = true
		})
	})
	assert.True(t, runCalled, "Global Run function did not execute")

	// Test global Test function
	var testCalled bool
	assert.NotPanics(t, func() {
		gone.Test(func(flag gone.TestFlag) {
			assert.NotNil(t, flag)
			testCalled = true
		})
	})
	assert.True(t, testCalled, "Global Test function did not execute")
}

func TestPreparer_RunTest(t *testing.T) {
	var testFuncCalled bool

	testFunc := func(flag gone.TestFlag) {
		assert.NotNil(t, flag)
		testFuncCalled = true
	}

	loadFn := func(core gone.Loader) error {
		return core.Load(&Worker{name: "test-worker"})
	}

	gone.RunTest(testFunc, loadFn)

	assert.True(t, testFuncCalled, "RunTest function was not called")
}

func TestPreparer_PrepareWithLoads(t *testing.T) {
	loadFn1 := func(core gone.Loader) error {
		return core.Load(&Worker{name: "worker1"})
	}
	loadFn2 := func(core gone.Loader) error {
		return core.Load(&Worker{name: "worker2"})
	}

	assert.NotPanics(t, func() {
		preparer := gone.Prepare(loadFn1, loadFn2)
		assert.NotNil(t, preparer)
	})

	// Test prepare with error load function
	errorLoadFn := func(core gone.Loader) error {
		return errors.New("load error")
	}

	assert.Panics(t, func() {
		gone.Prepare(errorLoadFn)
	})
}

func TestPreparer_ServeGlobal(t *testing.T) {
	// Test global Serve function
	go func() {
		time.Sleep(100 * time.Millisecond)
		gone.Default.End()
	}()

	assert.NotPanics(t, func() {
		gone.Serve()
	})
}
