package gone

import (
	"reflect"
	"testing"
)

func TestOption_Apply(t *testing.T) {
	tests := []struct {
		name  string
		apply func(c *coffin) error

		wantErr bool
	}{
		{
			name:    "Nil apply function",
			apply:   nil,
			wantErr: false,
		},
		{
			name: "Valid apply function",
			apply: func(c *coffin) error {
				c.name = "test"
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := option{apply: tt.apply}
			c := &coffin{}
			err := opt.Apply(c)

			if (err != nil) != tt.wantErr {
				t.Errorf("option.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrder(t *testing.T) {
	tests := []struct {
		name      string
		order     int
		wantOrder int
	}{
		{"Positive order", 42, 42},
		{"Zero order", 0, 0},
		{"Negative order", -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &coffin{}
			opt := Order(tt.order)

			if err := opt.Apply(c); err != nil {
				t.Errorf("Order(%v).Apply() error = %v", tt.order, err)
			}

			if c.order != tt.wantOrder {
				t.Errorf("Order(%v) got order = %v, want %v", tt.order, c.order, tt.wantOrder)
			}
		})
	}
}

func TestName(t *testing.T) {
	tests := []struct {
		name     string
		setName  string
		wantName string
	}{
		{"Normal name", "test-component", "test-component"},
		{"Empty name", "", ""},
		{"Special characters", "test@123_-.", "test@123_-."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &coffin{}
			opt := Name(tt.setName)

			if err := opt.Apply(c); err != nil {
				t.Errorf("GonerName(%q).Apply() error = %v", tt.setName, err)
			}

			if c.name != tt.wantName {
				t.Errorf("GonerName(%q) got name = %q, want %q", tt.setName, c.name, tt.wantName)
			}
		})
	}
}

func TestOnlyForName(t *testing.T) {
	c := &coffin{}
	opt := OnlyForName()

	if err := opt.Apply(c); err != nil {
		t.Errorf("OnlyForName().Apply() error = %v", err)
	}

	if !c.onlyForName {
		t.Error("OnlyForName() did not set onlyForName to true")
	}
}

func TestForceReplace(t *testing.T) {
	c := &coffin{}
	opt := ForceReplace()

	if err := opt.Apply(c); err != nil {
		t.Errorf("ForceReplace().Apply() error = %v", err)
	}

	if !c.forceReplace {
		t.Error("ForceReplace() did not set forceReplace to true")
	}
}

func TestIsDefault(t *testing.T) {
	type testStruct struct{}
	tests := []struct {
		name      string
		input     []any
		wantPanic bool
	}{
		{
			name:      "Valid pointer",
			input:     []any{&testStruct{}},
			wantPanic: false,
		},
		{
			name:      "Multiple valid pointers",
			input:     []any{&testStruct{}, &testStruct{}},
			wantPanic: false,
		},
		{
			name:      "Non-pointer value should panic",
			input:     []any{testStruct{}},
			wantPanic: true,
		},
		{
			name:      "Empty input",
			input:     []any{},
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("IsDefault() panic = %v, wantPanic = %v", r != nil, tt.wantPanic)
				}
			}()

			c := &coffin{
				defaultTypeMap: make(map[reflect.Type]bool),
			}
			opt := IsDefault(tt.input...)
			err := opt.Apply(c)

			if err != nil && !tt.wantPanic {
				t.Errorf("IsDefault().Apply() error = %v", err)
			}

			// 验证类型是否被正确标记为默认
			if !tt.wantPanic && len(tt.input) > 0 {
				for _, input := range tt.input {
					if ptr, ok := input.(*testStruct); ok {
						if !c.defaultTypeMap[reflect.TypeOf(ptr).Elem()] {
							t.Errorf("Type %v was not marked as default", reflect.TypeOf(ptr).Elem())
						}
					}
				}
			}
		})
	}
}

func TestPriorityOptions(t *testing.T) {
	tests := []struct {
		name      string
		option    Option
		wantOrder int
	}{
		{
			name:      "High priority",
			option:    HighStartPriority(),
			wantOrder: -100,
		},
		{
			name:      "Medium priority",
			option:    MediumStartPriority(),
			wantOrder: 0,
		},
		{
			name:      "Low priority",
			option:    LowStartPriority(),
			wantOrder: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &coffin{}
			err := tt.option.Apply(c)

			if err != nil {
				t.Errorf("%s().Apply() error = %v", tt.name, err)
			}

			if c.order != tt.wantOrder {
				t.Errorf("%s() got order = %v, want %v", tt.name, c.order, tt.wantOrder)
			}
		})
	}
}

func TestLazyFill(t *testing.T) {
	c := &coffin{}
	opt := LazyFill()

	if err := opt.Apply(c); err != nil {
		t.Errorf("LazyFill().Apply() error = %v", err)
	}

	if !c.lazyFill {
		t.Error("LazyFill() did not set lazyFill to true")
	}
}
