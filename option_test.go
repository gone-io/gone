package gone

import (
	"testing"
)

func TestOption_Apply(t *testing.T) {
	tests := []struct {
		name    string
		apply   func(c *coffin) error
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

func TestIsDefault(t *testing.T) {
	c := &coffin{}
	opt := IsDefault()

	if err := opt.Apply(c); err != nil {
		t.Errorf("IsDefault().Apply() error = %v", err)
	}

	if !c.isDefault {
		t.Error("IsDefault() did not set isDefault to true")
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
				t.Errorf("Name(%q).Apply() error = %v", tt.setName, err)
			}

			if c.name != tt.wantName {
				t.Errorf("Name(%q) got name = %q, want %q", tt.setName, c.name, tt.wantName)
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

func TestOptionCombination(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
		want *coffin
	}{
		{
			name: "Multiple options",
			opts: []Option{
				Name("test"),
				Order(42),
				IsDefault(),
				OnlyForName(),
				ForceReplace(),
			},
			want: &coffin{
				name:         "test",
				order:        42,
				isDefault:    true,
				onlyForName:  true,
				forceReplace: true,
			},
		},
		{
			name: "Override name",
			opts: []Option{
				Name("first"),
				Name("second"),
			},
			want: &coffin{
				name: "second",
			},
		},
		{
			name: "Override order",
			opts: []Option{
				Order(1),
				Order(2),
			},
			want: &coffin{
				order: 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &coffin{}

			// Apply all options
			for _, opt := range tt.opts {
				if err := opt.Apply(c); err != nil {
					t.Errorf("Option.Apply() error = %v", err)
					return
				}
			}

			// Check all relevant fields
			if c.name != tt.want.name {
				t.Errorf("got name = %q, want %q", c.name, tt.want.name)
			}
			if c.order != tt.want.order {
				t.Errorf("got order = %v, want %v", c.order, tt.want.order)
			}
			if c.isDefault != tt.want.isDefault {
				t.Errorf("got isDefault = %v, want %v", c.isDefault, tt.want.isDefault)
			}
			if c.onlyForName != tt.want.onlyForName {
				t.Errorf("got onlyForName = %v, want %v", c.onlyForName, tt.want.onlyForName)
			}
			if c.forceReplace != tt.want.forceReplace {
				t.Errorf("got forceReplace = %v, want %v", c.forceReplace, tt.want.forceReplace)
			}
		})
	}
}
