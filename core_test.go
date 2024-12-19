package gone

import (
	"errors"
	"reflect"
	"testing"
)

// Mock components for testing
type MockComponent struct {
	Flag
	Dependency *MockDependency `gone:"*"`
	Named      *MockNamed      `gone:"named-dep"`
}

type MockDependency struct {
	Flag
}

type MockNamed struct {
	Flag
}

func (m *MockNamed) GonerName() string {
	return "named-dep"
}

type MockProvider struct {
	Flag
	returnVal any
	returnErr error
}

func (m *MockProvider) Provide(conf string, t reflect.Type) (any, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	return m.returnVal, nil
}

func (m *MockProvider) GonerName() string {
	return "mock-provider"
}

type MockInitiator struct {
	Flag
	initCalled      bool
	beforeInitError error
	initError       error
}

func (m *MockInitiator) BeforeInit() error {
	return m.beforeInitError
}

func (m *MockInitiator) Init() error {
	m.initCalled = true
	return m.initError
}

func TestNewCore(t *testing.T) {
	core := NewCore()

	if core == nil {
		t.Error("NewCore() returned nil")
		return
	}

	if core.nameMap == nil {
		t.Error("nameMap not initialized")
	}

	if core.typeProviderMap == nil {
		t.Error("typeProviderMap not initialized")
	}

	if core.typeProviderDepMap == nil {
		t.Error("typeProviderDepMap not initialized")
	}
}

func TestCore_Load(t *testing.T) {
	tests := []struct {
		name    string
		goner   Goner
		options []Option
		wantErr bool
	}{
		{
			name:    "Basic component",
			goner:   &MockComponent{},
			options: nil,
			wantErr: false,
		},
		{
			name:    "Named component",
			goner:   &MockNamed{},
			options: nil,
			wantErr: false,
		},
		{
			name:  "Component with name option",
			goner: &MockComponent{},
			options: []Option{
				Name("custom-name"),
			},
			wantErr: false,
		},
		{
			name:  "Duplicate name without force replace",
			goner: &MockNamed{},
			options: []Option{
				Name("duplicate"),
			},
			wantErr: true,
		},
		{
			name:  "Duplicate name with force replace",
			goner: &MockNamed{},
			options: []Option{
				Name("duplicate"),
				ForceReplace(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			_ = core.Load(&MockNamed{}, Name("duplicate"))

			err := core.Load(tt.goner, tt.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Core.Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type Circular1 struct {
	Flag
	Dep *Circular2 `gone:"*"`
}

func (s *Circular1) Init() {

}

type Circular2 struct {
	Flag
	Dep *Circular1 `gone:"*"`
}

func (s *Circular2) Init() {

}

func TestCore_Fill(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Core)
		wantErr bool
	}{
		{
			name: "Basic dependency injection",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
				_ = core.Load(&MockNamed{})
				_ = core.Load(&MockComponent{})
			},
			wantErr: false,
		},
		{
			name: "Missing dependency",
			setup: func(core *Core) {
				_ = core.Load(&MockComponent{})
			},
			wantErr: true,
		},
		{
			name: "Circular dependency",
			setup: func(core *Core) {

				_ = core.Load(&Circular1{})
				_ = core.Load(&Circular2{})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			_ = core.Load(defaultLog)
			_ = core.Load(&ConfigProvider{})
			_ = core.Load(&EnvConfigure{}, Name("configure"), IsDefault(new(Configure)), OnlyForName())
			tt.setup(core)
			err := core.Install()
			if (err != nil) != tt.wantErr {
				t.Errorf("Core.Fill() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCore_Check(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*Core)
		wantErr      bool
		wantOrderLen int
	}{
		{
			name: "Valid dependency order",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
				_ = core.Load(&MockComponent{})
				_ = core.Load(&MockNamed{})
			},
			wantErr:      false,
			wantOrderLen: 14,
		},
		{
			name: "Circular dependency",
			setup: func(core *Core) {
				type Circular1 struct {
					Flag
					Dep *Circular2 `gone:"*"`
				}
				type Circular2 struct {
					Flag
					Dep *Circular1 `gone:"*"`
				}
				_ = core.Load(&Circular1{})
				_ = core.Load(&Circular2{})
			},
			wantErr:      true,
			wantOrderLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			_ = core.Load(defaultLog)
			_ = core.Load(&ConfigProvider{})
			_ = core.Load(&EnvConfigure{}, Name("configure"), IsDefault(new(Configure)), OnlyForName())
			tt.setup(core)
			orders, err := core.Check()
			if (err != nil) != tt.wantErr {
				t.Errorf("Core.Check() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(orders) != tt.wantOrderLen {
				t.Errorf("Core.Check() order length = %v, want %v", len(orders), tt.wantOrderLen)
			}
		})
	}
}

func TestCore_InjectFuncParameters(t *testing.T) {
	type testStruct struct {
		Flag
		Value *MockDependency `gone:"*"`
	}

	tests := []struct {
		name         string
		setup        func(*Core)
		fn           any
		injectBefore FuncInjectHook
		injectAfter  FuncInjectHook
		wantErr      bool
	}{
		{
			name: "Basic function injection",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
			},
			fn:      func(dep *MockDependency) {},
			wantErr: false,
		},
		{
			name: "Struct parameter injection",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
			},
			fn:      func(s testStruct) {},
			wantErr: false,
		},
		{
			name:    "Invalid function",
			setup:   func(core *Core) {},
			fn:      "not a function",
			wantErr: true,
		},
		{
			name:  "With inject hooks",
			setup: func(core *Core) {},
			fn:    func(s string) {},
			injectBefore: func(pt reflect.Type, i int, injected bool) any {
				return "injected value"
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)
			_, err := core.InjectFuncParameters(tt.fn, tt.injectBefore, tt.injectAfter)
			if (err != nil) != tt.wantErr {
				t.Errorf("Core.InjectFuncParameters() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCore_InjectWrapFunc(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*Core)
		fn           any
		injectBefore FuncInjectHook
		injectAfter  FuncInjectHook
		wantResults  []any
		wantErr      bool
	}{
		{
			name: "Function with return values",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
			},
			fn: func(dep *MockDependency) string {
				return "test"
			},
			wantResults: []any{"test"},
			wantErr:     false,
		},
		{
			name: "Function with error",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
			},
			fn: func(dep *MockDependency) error {
				return errors.New("test error")
			},
			wantResults: []any{errors.New("test error")},
			wantErr:     false,
		},
		{
			name: "Function with multiple returns",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
			},
			fn: func(dep *MockDependency) (string, error) {
				return "test", nil
			},
			wantResults: []any{"test", nil},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			wrapper, err := core.InjectWrapFunc(tt.fn, tt.injectBefore, tt.injectAfter)
			if (err != nil) != tt.wantErr {
				t.Errorf("Core.InjectWrapFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				results := wrapper()
				if !reflect.DeepEqual(results, tt.wantResults) {
					t.Errorf("Wrapped function results = %v, want %v", results, tt.wantResults)
				}
			}
		})
	}
}
