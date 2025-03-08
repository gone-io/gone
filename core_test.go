package gone

import (
	"errors"
	"fmt"
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

type MockBeforeInitNoError struct {
	Flag
	beforeInitCalled bool
}

func (m *MockBeforeInitNoError) BeforeInit() {
	m.beforeInitCalled = true
}

type MockStructFieldInjector struct {
	Flag
}

func (m *MockStructFieldInjector) GonerName() string {
	return "field-injector"
}

func (m *MockStructFieldInjector) Inject(conf string, field reflect.StructField, v reflect.Value) error {
	if conf == "error" {
		return fmt.Errorf("injection error")
	}
	v.Set(reflect.ValueOf("injected value"))
	return nil
}

type StructWithUnexportedField struct {
	Flag
	dep    *MockDependency `gone:"*"`
	Public *MockDependency `gone:"*"`
}

// Test struct for provider error cases
type ErrorProvider struct {
	Flag
	returnErr error
}

func (e *ErrorProvider) Provide(conf string, t reflect.Type) (any, error) {
	return nil, e.returnErr
}

func (e *ErrorProvider) GonerName() string {
	return "error-provider"
}

// Test struct for slice injection
type SliceContainer struct {
	Flag
	Deps []*MockDependency `gone:"*"`
}

// Test struct for invalid field types
type InvalidFieldType struct {
	Flag
	Channel chan int `gone:"*"` // Should fail
}

// Test struct for BeforeInit and Init combinations
type MockInitCombinations struct {
	Flag
	beforeInitCalled bool
	initCalled       bool
	shouldError      bool
}

func (m *MockInitCombinations) BeforeInit() error {
	m.beforeInitCalled = true
	if m.shouldError {
		return fmt.Errorf("before init error")
	}
	return nil
}

func (m *MockInitCombinations) Init() error {
	m.initCalled = true
	if m.shouldError {
		return fmt.Errorf("init error")
	}
	return nil
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
		{
			name: "BeforeInitNoError implementation",
			setup: func(core *Core) {
				mock := &MockBeforeInitNoError{}
				_ = core.Load(mock)
			},
			wantErr: false,
		},
		{
			name: "Unexported field injection",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
				_ = core.Load(&StructWithUnexportedField{})
			},
			wantErr: false,
		},
		{
			name: "StructFieldInjector success",
			setup: func(core *Core) {
				_ = core.Load(&MockStructFieldInjector{})
				type TestStruct struct {
					Flag
					Value string `gone:"field-injector"`
				}
				_ = core.Load(&TestStruct{})
			},
			wantErr: false,
		},
		{
			name: "StructFieldInjector error",
			setup: func(core *Core) {
				_ = core.Load(&MockStructFieldInjector{})
				type TestStruct struct {
					Flag
					Value string `gone:"field-injector-error"`
				}
				_ = core.Load(&TestStruct{})
			},
			wantErr: true,
		},
		{
			name: "Provider with invalid return type",
			setup: func(core *Core) {
				_ = core.Load(&MockProvider{
					returnVal: "invalid",
				})
				type TestStruct struct {
					Flag
					Value *MockDependency `gone:"mock-provider"`
				}
				_ = core.Load(&TestStruct{})
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

func TestCore_GetGonerByName(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Core)
		findName string
		want     bool
	}{
		{
			name: "Find existing named component",
			setup: func(core *Core) {
				_ = core.Load(&MockNamed{})
			},
			findName: "named-dep",
			want:     true,
		},
		{
			name: "Component not found",
			setup: func(core *Core) {
				_ = core.Load(&MockComponent{})
			},
			findName: "non-existent",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			got := core.GetGonerByName(tt.findName)
			if (got != nil) != tt.want {
				t.Errorf("Core.GetGonerByName() = %v, want %v", got != nil, tt.want)
			}
		})
	}
}

func TestCore_GetGonerByType(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Core)
		typ   reflect.Type
		want  bool
	}{
		{
			name: "Find existing type",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
			},
			typ:  reflect.TypeOf(&MockDependency{}),
			want: true,
		},
		{
			name: "Multiple implementations with default",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{}, IsDefault())
				_ = core.Load(&MockDependency{})
			},
			typ:  reflect.TypeOf(&MockDependency{}),
			want: true,
		},
		{
			name:  "Type not found",
			setup: func(core *Core) {},
			typ:   reflect.TypeOf(&MockComponent{}),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			got := core.GetGonerByType(tt.typ)
			if (got != nil) != tt.want {
				t.Errorf("Core.GetGonerByType() = %v, want %v", got != nil, tt.want)
			}
		})
	}
}

func TestCore_InjectStruct(t *testing.T) {
	type testStruct struct {
		Flag
		Dep *MockDependency `gone:"*"`
	}

	tests := []struct {
		name    string
		setup   func(*Core)
		target  any
		wantErr bool
	}{
		{
			name: "Valid struct injection",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
			},
			target:  &testStruct{},
			wantErr: false,
		},
		{
			name:    "Non-pointer target",
			setup:   func(core *Core) {},
			target:  testStruct{},
			wantErr: true,
		},
		{
			name:    "Non-struct pointer",
			setup:   func(core *Core) {},
			target:  new(string),
			wantErr: true,
		},
		{
			name:    "Missing dependency",
			setup:   func(core *Core) {},
			target:  &testStruct{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			err := core.InjectStruct(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Core.InjectStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCore_Loaded(t *testing.T) {
	core := NewCore()
	key := GenLoaderKey()

	// First call should return false
	if core.Loaded(key) {
		t.Error("First call to Loaded() should return false")
	}

	// Second call should return true
	if !core.Loaded(key) {
		t.Error("Second call to Loaded() should return true")
	}
}

func TestCore_Provide(t *testing.T) {
	type testSlice []*MockDependency

	tests := []struct {
		name    string
		setup   func(*Core)
		typ     reflect.Type
		tagConf string
		want    bool
		wantErr bool
	}{
		{
			name: "Provide single component",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
			},
			typ:     reflect.TypeOf(&MockDependency{}),
			tagConf: "",
			want:    true,
			wantErr: false,
		},
		{
			name: "Provide slice of components",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
				_ = core.Load(&MockDependency{})
			},
			typ:     reflect.TypeOf(testSlice{}),
			tagConf: "",
			want:    true,
			wantErr: false,
		},
		{
			name: "Provider returns error",
			setup: func(core *Core) {
				_ = core.Load(&MockProvider{returnErr: fmt.Errorf("provider error")})
			},
			typ:     reflect.TypeOf(&MockDependency{}),
			tagConf: "",
			want:    false,
			wantErr: true,
		},
		{
			name: "Provider returns incompatible type",
			setup: func(core *Core) {
				_ = core.Load(&MockProvider{
					returnVal: "invalid string instead of MockDependency",
				})
			},
			typ:     reflect.TypeOf(&MockDependency{}),
			tagConf: "mock-provider",
			want:    false,
			wantErr: true,
		},
		{
			name: "Slice with mixed sources",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
				_ = core.Load(&MockProvider{
					returnVal: &MockDependency{},
				})
			},
			typ:     reflect.TypeOf([]*MockDependency{}),
			tagConf: "",
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			got, err := core.Provide(tt.tagConf, tt.typ)
			if (err != nil) != tt.wantErr {
				t.Errorf("Core.Provide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got != nil) != tt.want {
				t.Errorf("Core.Provide() = %v, want %v", got != nil, tt.want)
			}
		})
	}
}

func TestCore_InjectStruct_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		target  interface{}
		setup   func(*Core)
		wantErr bool
	}{
		{
			name: "Nil pointer",
			target: (*struct {
				Flag
				Dep *MockDependency `gone:"*"`
			})(nil),
			setup:   func(core *Core) {},
			wantErr: true,
		},
		{
			name:    "Non-struct pointer",
			target:  new(string),
			setup:   func(core *Core) {},
			wantErr: true,
		},
		{
			name: "Invalid tag configuration",
			target: &struct {
				Flag
				Dep *MockDependency `gone:"invalid,config"`
			}{},
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			err := core.InjectStruct(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("InjectStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCore_Load_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		goner   Goner
		options []Option
		setup   func(*Core)
		wantErr bool
	}{
		{
			name:    "Nil goner",
			goner:   nil,
			wantErr: true,
		},
		{
			name:  "Provider with same type and force replace",
			goner: &MockProvider{},
			options: []Option{
				ForceReplace(),
			},
			setup: func(core *Core) {
				_ = core.Load(&MockProvider{})
			},
			wantErr: false,
		},
		{
			name:  "Provider with same type without force replace",
			goner: &MockProvider{},
			setup: func(core *Core) {
				_ = core.Load(&MockProvider{})
			},
			wantErr: true,
		},
		{
			name:  "Named component with force replace",
			goner: &MockNamed{},
			options: []Option{
				Name("test-name"),
				ForceReplace(),
			},
			setup: func(core *Core) {
				_ = core.Load(&MockComponent{}, Name("test-name"))
			},
			wantErr: false,
		},
		{
			name:  "Named component without force replace",
			goner: &MockNamed{},
			options: []Option{
				Name("test-name"),
			},
			setup: func(core *Core) {
				_ = core.Load(&MockComponent{}, Name("test-name"))
			},
			wantErr: true,
		},
		{
			name:  "Component with only for name",
			goner: &MockComponent{},
			options: []Option{
				Name("test-name"),
				OnlyForName(),
			},
			wantErr: false,
		},
		{
			name:  "Component with lazy fill",
			goner: &MockComponent{},
			options: []Option{
				LazyFill(),
			},
			wantErr: false,
		},
		{
			name:  "Component with order",
			goner: &MockComponent{},
			options: []Option{
				Order(100),
			},
			wantErr: false,
		},
		{
			name:  "Component with priority options",
			goner: &MockComponent{},
			options: []Option{
				HighStartPriority(),
				MediumStartPriority(),
				LowStartPriority(),
			},
			wantErr: false,
		},
		{
			name:  "Component with multiple options",
			goner: &MockComponent{},
			options: []Option{
				Name("test-name"),
				OnlyForName(),
				LazyFill(),
				Order(100),
				IsDefault(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantErr {
						t.Errorf("Core.Load() unexpected panic = %v", r)
					}
				}
			}()

			core := NewCore()
			if tt.setup != nil {
				tt.setup(core)
			}

			err := core.Load(tt.goner, tt.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Core.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify the options were applied correctly
				if len(tt.options) > 0 {
					var co *coffin
					for _, c := range core.coffins {
						if c.goner == tt.goner {
							co = c
							break
						}
					}
					if co == nil {
						t.Error("Core.Load() coffin not found after loading")
						return
					}
				}
			}
		})
	}
}

func TestCore_Provide_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Core)
		typ     reflect.Type
		tagConf string
		wantErr bool
	}{
		{
			name: "Provider returns error",
			setup: func(core *Core) {
				_ = core.Load(&ErrorProvider{
					returnErr: fmt.Errorf("provider error"),
				})
			},
			typ:     reflect.TypeOf(&MockDependency{}),
			tagConf: "error-provider",
			wantErr: true,
		},
		{
			name: "Slice with provider error",
			setup: func(core *Core) {
				_ = core.Load(&ErrorProvider{
					returnErr: fmt.Errorf("provider error"),
				})
			},
			typ:     reflect.TypeOf([]*MockDependency{}),
			tagConf: "error-provider",
			wantErr: false, // Slice should still be returned even if provider fails
		},
		{
			name: "Multiple providers for slice",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
				_ = core.Load(&MockDependency{})
				_ = core.Load(&MockProvider{
					returnVal: &MockDependency{},
				})
			},
			typ:     reflect.TypeOf([]*MockDependency{}),
			tagConf: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			got, err := core.Provide(tt.tagConf, tt.typ)
			if (err != nil) != tt.wantErr {
				t.Errorf("Core.Provide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("Core.Provide() returned nil but expected a value")
			}
		})
	}
}

func TestCore_InjectFuncParameters_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		fn          interface{}
		setup       func(*Core)
		injectHooks struct {
			before FuncInjectHook
			after  FuncInjectHook
		}
		wantErr bool
	}{
		{
			name: "Non-function input",
			fn:   "not a function",
			setup: func(core *Core) {
			},
			wantErr: true,
		},
		{
			name: "Function with unsupported parameter type",
			fn: func(ch chan int) {
			},
			setup: func(core *Core) {
			},
			wantErr: true,
		},
		{
			name: "Before hook provides value",
			fn: func(s string) {
			},
			setup: func(core *Core) {
			},
			injectHooks: struct {
				before FuncInjectHook
				after  FuncInjectHook
			}{
				before: func(pt reflect.Type, i int, injected bool) any {
					return "injected value"
				},
			},
			wantErr: false,
		},
		{
			name: "After hook provides value",
			fn: func(s string) {
			},
			setup: func(core *Core) {
			},
			injectHooks: struct {
				before FuncInjectHook
				after  FuncInjectHook
			}{
				after: func(pt reflect.Type, i int, injected bool) any {
					return "injected value"
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			_, err := core.InjectFuncParameters(tt.fn, tt.injectHooks.before, tt.injectHooks.after)
			if (err != nil) != tt.wantErr {
				t.Errorf("InjectFuncParameters() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCore_InjectWrapFunc_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		fn      interface{}
		setup   func(*Core)
		wantErr bool
	}{
		{
			name: "Function returning nil interface",
			fn: func() error {
				return nil
			},
			setup:   func(core *Core) {},
			wantErr: false,
		},
		{
			name: "Function returning multiple values",
			fn: func() (string, error, int) {
				return "test", nil, 42
			},
			setup:   func(core *Core) {},
			wantErr: false,
		},
		{
			name: "Function with invalid parameter injection",
			fn: func(ch chan int) {
			},
			setup:   func(core *Core) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			wrapper, err := core.InjectWrapFunc(tt.fn, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("InjectWrapFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && wrapper == nil {
				t.Error("InjectWrapFunc() returned nil wrapper but expected a function")
			}

			if !tt.wantErr {
				results := wrapper()
				if results == nil {
					t.Error("Wrapper function returned nil results")
				}
			}
		})
	}
}

func TestCore_Fill_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Core)
		wantErr bool
	}{
		{
			name: "Fill with invalid tag format",
			setup: func(core *Core) {
				type InvalidTag struct {
					Flag
					Dep *MockDependency `gone:"invalid:tag:format"`
				}
				_ = core.Load(&InvalidTag{})
			},
			wantErr: true,
		},
		{
			name: "Fill with non-existent provider",
			setup: func(core *Core) {
				type MissingProvider struct {
					Flag
					Dep *MockDependency `gone:"missing-provider"`
				}
				_ = core.Load(&MissingProvider{})
			},
			wantErr: true,
		},
		{
			name: "Fill with incompatible provider type",
			setup: func(core *Core) {
				_ = core.Load(&MockProvider{
					returnVal: "string instead of MockDependency",
				})
				type IncompatibleType struct {
					Flag
					Dep *MockDependency `gone:"mock-provider"`
				}
				_ = core.Load(&IncompatibleType{})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)
			err := core.Install()
			if (err != nil) != tt.wantErr {
				t.Errorf("Core.Fill() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCore_Init_Combinations(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Core, *MockInitCombinations)
		verify  func(*testing.T, *MockInitCombinations)
		wantErr bool
	}{
		{
			name: "All init methods success",
			setup: func(core *Core, mock *MockInitCombinations) {
				mock.shouldError = false
				_ = core.Load(mock)
			},
			verify: func(t *testing.T, mock *MockInitCombinations) {
				if !mock.beforeInitCalled {
					t.Error("BeforeInit was not called")
				}
				if !mock.initCalled {
					t.Error("Init was not called")
				}
			},
			wantErr: false,
		},
		{
			name: "BeforeInit returns error",
			setup: func(core *Core, mock *MockInitCombinations) {
				mock.shouldError = true
				_ = core.Load(mock)
			},
			verify: func(t *testing.T, mock *MockInitCombinations) {
				if !mock.beforeInitCalled {
					t.Error("BeforeInit was not called")
				}

				if mock.initCalled {
					t.Error("Init should not have been called")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			mock := &MockInitCombinations{}
			tt.setup(core, mock)

			err := core.Install()
			if (err != nil) != tt.wantErr {
				t.Errorf("Core.Install() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.verify != nil {
				tt.verify(t, mock)
			}
		})
	}
}

func TestCore_GetGonerByType_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Core)
		typ     reflect.Type
		wantNil bool
	}{
		{
			name: "Multiple implementations without default",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
				_ = core.Load(&MockDependency{})
			},
			typ:     reflect.TypeOf(&MockDependency{}),
			wantNil: false, // Should return first one
		},
		{
			name: "Multiple implementations with default",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
				_ = core.Load(&MockDependency{}, IsDefault())
			},
			typ:     reflect.TypeOf(&MockDependency{}),
			wantNil: false,
		},
		{
			name: "Only for name implementation",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{}, OnlyForName())
			},
			typ:     reflect.TypeOf(&MockDependency{}),
			wantNil: true,
		},
		{
			name: "Interface type",
			setup: func(core *Core) {
				type TestInterface interface {
					Test()
				}
				type TestImpl struct {
					Flag
				}
				_ = core.Load(&TestImpl{})
			},
			typ:     reflect.TypeOf((*error)(nil)).Elem(), // Use error interface as example
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			got := core.GetGonerByType(tt.typ)
			if (got == nil) != tt.wantNil {
				t.Errorf("GetGonerByType() = %v, want nil: %v", got, tt.wantNil)
			}
		})
	}
}

func TestCore_SafeExecute(t *testing.T) {
	tests := []struct {
		name    string
		fn      func() error
		wantErr bool
	}{
		{
			name: "Normal execution",
			fn: func() error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "Function returns error",
			fn: func() error {
				return fmt.Errorf("test error")
			},
			wantErr: true,
		},
		{
			name: "Function panics",
			fn: func() error {
				panic("test panic")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SafeExecute(tt.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeExecute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestActionType_String(t *testing.T) {
	tests := []struct {
		name   string
		action actionType
		want   string
	}{
		{
			name:   "fill",
			action: fillAction,
			want:   "fill fields",
		},
		{
			name:   "init",
			action: initAction,
			want:   "initialize",
		},
		{
			name:   "unknown",
			action: actionType(0),
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.action.String() != tt.want {
				t.Errorf("ActionType.String() = %v, want %v", tt.action.String(), tt.want)
			}
		})
	}
}

type testOption struct {
}

func (t testOption) Apply(c *coffin) error {
	return ToError("apply-error")
}

func TestLoadWithOptionApplyError(t *testing.T) {
	core := NewCore()
	type Test struct {
		Flag
	}
	err := core.Load(&Test{}, testOption{})
	if err == nil {
		t.Error("Expected error, but got nil")
	}
}

func TestLoadWithReplaceOption(t *testing.T) {
	core := NewCore()
	type Test struct {
		Flag
	}

	var test Test

	provider := WrapFunctionProvider(func(tagConf string, in struct{}) (Test, error) {
		return test, ToError("test error")
	})

	_ = core.Load(provider)
	err := core.Load(provider)
	if err == nil {
		t.Error("Expected error, but got nil")
	}

	provider2 := WrapFunctionProvider(func(tagConf string, in struct{}) (Test, error) {
		return test, ToError("test error")
	})
	err = core.Load(provider2)
	if err == nil {
		t.Error("Expected error, but got nil")
	}
	err = core.Load(provider2, ForceReplace())
	if err != nil {
		t.Error("Expected no error, but got:", err)
	}

}

type errInit struct {
	Flag
}

func (e *errInit) Init() {
	panic("error")
}

func TestInstallWithError(t *testing.T) {
	core := NewCore()
	err := core.Load(&errInit{})
	if err != nil {
		t.Error("Expected no error, but got:", err)
	}
	err = core.Install()
	if err == nil {
		t.Error("Expected error, but got nil")
	}
}

func TestFillOne(t *testing.T) {
	core := NewCore()

	var test string

	c := newCoffin(&test)

	err := core.fillOne(c)
	if err == nil {
		t.Error("Expected error")
	}

	type testStruct struct {
		Flag
		dep  *Core  `gone:""`
		test string `gone:"*"`
	}

	co := newCoffin(&testStruct{})

	err = core.fillOne(co)
	if err == nil {
		t.Error("Expected error")
	}
	type testStruct2 struct {
		test string `gone:"okk"`
	}

	co2 := newCoffin(&testStruct2{})

	_ = core.Load(&testStruct{}, Name("okk"))
	err = core.fillOne(co2)
	if err == nil {
		t.Error("Expected error")
	}

}

type initEr struct {
	Flag
}

func (i *initEr) Init() error {
	return ToError("init-error")
}

func TestInitOneError(t *testing.T) {
	core := NewCore()
	c := newCoffin(&initEr{})

	err := core.initOne(c)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestProvide(t *testing.T) {
	type Test struct {
	}

	var test Test

	provider := WrapFunctionProvider(func(tagConf string, in struct{}) (Test, error) {
		return test, ToError("test error")
	})

	var test2 []Test

	NewApp(func(loader Loader) error {
		return loader.Load(provider)
	}).
		Test(func(core *Core) {
			_, err := core.Provide("", reflect.TypeOf(test2))
			if err == nil {
				t.Errorf("Expected error, got nil")
			}
		})

	provider2 := WrapFunctionProvider(func(tagConf string, in struct{}) (Test, error) {
		return test, nil
	})

	NewApp(func(loader Loader) error {
		return loader.Load(provider2)
	}).
		Test(func(core *Core) {
			_, err := core.Provide("", reflect.TypeOf(test2))
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
}

func TestInjectFuncParametersWithStructParameter(t *testing.T) {
	NewApp().
		Test(func(core *Core) {
			_, err := core.InjectFuncParameters(func(in *struct {
				core *Core `gone:"*"`
			}) {
			}, nil, nil)
			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})

	NewApp().
		Test(func(core *Core) {
			_, err := core.InjectFuncParameters(func(in *struct {
				core string `gone:"*"`
			}) {
			}, nil, nil)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}

			_, err = core.InjectFuncParameters(func(in struct {
				core string `gone:"*"`
			}) {
			}, nil, nil)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}
		})
}
