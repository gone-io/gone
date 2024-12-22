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
