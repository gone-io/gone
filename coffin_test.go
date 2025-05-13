package gone

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestCoffinListSort(t *testing.T) {
	tests := []struct {
		name     string
		coffins  []*coffin
		expected []*coffin
	}{
		{
			name: "Sort by order - ascending",
			coffins: []*coffin{
				{name: "C", order: 3},
				{name: "A", order: 1},
				{name: "B", order: 2},
			},
			expected: []*coffin{
				{name: "A", order: 1},
				{name: "B", order: 2},
				{name: "C", order: 3},
			},
		},
		{
			name: "Sort with same orders",
			coffins: []*coffin{
				{name: "A", order: 1},
				{name: "B", order: 1},
				{name: "C", order: 1},
			},
			expected: []*coffin{
				{name: "A", order: 1},
				{name: "B", order: 1},
				{name: "C", order: 1},
			},
		},
		{
			name: "Sort with negative orders",
			coffins: []*coffin{
				{name: "C", order: 1},
				{name: "A", order: -2},
				{name: "B", order: -1},
			},
			expected: []*coffin{
				{name: "A", order: -2},
				{name: "B", order: -1},
				{name: "C", order: 1},
			},
		},
		{
			name:     "Empty slice",
			coffins:  []*coffin{},
			expected: []*coffin{},
		},
		{
			name: "Single element",
			coffins: []*coffin{
				{name: "A", order: 1},
			},
			expected: []*coffin{
				{name: "A", order: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Sort method
			SortCoffins(tt.coffins)

			// Check if the result matches expected
			if !reflect.DeepEqual(tt.coffins, tt.expected) {
				t.Errorf("SortCoffins() got = %v, want %v", formatCoffins(tt.coffins), formatCoffins(tt.expected))
			}

			// Verify that the result is actually sorted
			for i := 1; i < len(tt.coffins); i++ {
				if tt.coffins[i-1].order > tt.coffins[i].order {
					t.Errorf("SortCoffins() result is not sorted at index %d: %v", i, formatCoffins(tt.coffins))
				}
			}
		})
	}
}

func TestCoffinListInterface(t *testing.T) {
	list := coffinList{
		{name: "B", order: 2},
		{name: "A", order: 1},
		{name: "C", order: 3},
	}

	// Test Len()
	if got := list.Len(); got != 3 {
		t.Errorf("coffinList.Len() = %v, want %v", got, 3)
	}

	// Test Less()
	tests := []struct {
		i, j     int
		expected bool
	}{
		{0, 1, false}, // 2 > 1, should be false
		{1, 2, true},  // 1 < 3, should be true
		{0, 2, true},  // 2 < 3, should be true
	}

	for _, tt := range tests {
		t.Run(formatLessTest(list[tt.i], list[tt.j]), func(t *testing.T) {
			if got := list.Less(tt.i, tt.j); got != tt.expected {
				t.Errorf("coffinList.Less(%v, %v) = %v, want %v",
					tt.i, tt.j, got, tt.expected)
			}
		})
	}

	// Test Swap()
	original := make([]*coffin, len(list))
	copy(original, list)

	list.Swap(0, 2)
	if list[0].order != original[2].order || list[2].order != original[0].order {
		t.Errorf("coffinList.Swap(0, 2) failed, got %v, want swapped %v",
			formatCoffins(list), formatCoffins(original))
	}
}

// Helper function to format coffins for error messages
func formatCoffins(coffins []*coffin) string {
	result := "["
	for i, c := range coffins {
		if i > 0 {
			result += ", "
		}
		result += "{" + c.name + ":" + string(rune('0'+c.order)) + "}"
	}
	result += "]"
	return result
}

// Helper function to format Less test description
func formatLessTest(a, b *coffin) string {
	return "Less " + a.name + "(" + string(rune('0'+a.order)) + ") " +
		b.name + "(" + string(rune('0'+b.order)) + ")"
}

type testInitiator struct {
	Flag
}

func (t *testInitiator) Init() error {
	return nil
}

type testInitiatorNoError struct {
	Flag
}

func (t *testInitiatorNoError) Init() {}

type testStructFieldInjector struct {
	Flag
}

func (t *testStructFieldInjector) GonerName() string {
	return "testInjector"
}

func (t *testStructFieldInjector) Inject(tagConf string, field reflect.StructField, fieldValue reflect.Value) error {
	return nil
}

type testNormalGoner struct {
	Flag
}

type testNamedProvider struct {
	Flag
}

func (t *testNamedProvider) GonerName() string {
	return "testProvider"
}

func (t *testNamedProvider) Provide(tagConf string, typ reflect.Type) (any, error) {
	return nil, nil
}

func TestNewCoffin_NeedInitBeforeUse(t *testing.T) {
	tests := []struct {
		name            string
		goner           Goner
		wantNeedInitUse bool
	}{
		{
			name:            "Initiator implementation",
			goner:           &testInitiator{},
			wantNeedInitUse: true,
		},
		{
			name:            "InitiatorNoError implementation",
			goner:           &testInitiatorNoError{},
			wantNeedInitUse: true,
		},
		{
			name:            "NamedProvider implementation",
			goner:           &testNamedProvider{},
			wantNeedInitUse: true,
		},
		{
			name:            "StructFieldInjector implementation",
			goner:           &testStructFieldInjector{},
			wantNeedInitUse: true,
		},
		{
			name:            "Normal Goner without special interfaces",
			goner:           &testNormalGoner{},
			wantNeedInitUse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			co := newCoffin(tt.goner)
			if co.needInitBeforeUse != tt.wantNeedInitUse {
				t.Errorf("newCoffin() needInitBeforeUse = %v, want %v",
					co.needInitBeforeUse, tt.wantNeedInitUse)
			}
		})
	}
}

type g1 struct{}

func (g1) Method() {

}

type i1 interface {
	Method()
}

func Test_coffin_CoundProvide(t *testing.T) {
	var x struct {
		g *g1
		i i1
	}

	of := reflect.TypeOf(&x).Elem()
	gField, _ := of.FieldByName("g")
	iField, _ := of.FieldByName("i")

	provider := WrapFunctionProvider(func(tagConf string, param struct{}) (*g1, error) {
		return nil, nil
	})

	c := newCoffin(&testNamedProvider{})

	c.defaultTypeMap[reflect.TypeOf(&g1{})] = true
	c.defaultTypeMap[reflect.TypeOf(new(i1)).Elem()] = true

	type args struct {
		t      reflect.Type
		byName bool
	}
	tests := []struct {
		name        string
		coffin      *coffin
		args        args
		errContains string
	}{
		{
			name:   "for compatible pointer",
			coffin: newCoffin(&g1{}),
			args: args{
				t:      gField.Type,
				byName: false,
			},
		},
		{
			name:   "for compatible interface",
			coffin: newCoffin(&g1{}),
			args: args{
				t:      iField.Type,
				byName: false,
			},
		},
		{
			name:   "for compatible pointer with provider",
			coffin: newCoffin(provider),
			args: args{
				t:      gField.Type,
				byName: false,
			},
		},
		{
			name:   "for compatible interface with provider",
			coffin: newCoffin(provider),
			args: args{
				t:      iField.Type,
				byName: false,
			},
		},
		{
			name:   "named provider with default pointer type",
			coffin: c,
			args: args{
				t:      gField.Type,
				byName: false,
			},
		},

		{
			name:   "named provider with default interface type",
			coffin: c,
			args: args{
				t:      iField.Type,
				byName: false,
			},
		},
		{
			name:   "named provider with byName = true",
			coffin: newCoffin(&testNamedProvider{}),
			args: args{
				t:      iField.Type,
				byName: true,
			},
		},
		{
			name:   "err",
			coffin: newCoffin(&testNamedProvider{}),
			args: args{
				t:      reflect.TypeOf(new(i1)).Elem(),
				byName: false,
			},
			errContains: `"Goner(name=testProvider)" cannot provide "github.com/gone-io/gone/v2.i1" value`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.coffin
			err := c.CoundProvide(tt.args.t, tt.args.byName)
			if tt.errContains == "" {
				if err != nil {
					t.Errorf("CoundProvide() error = %v, errContains %v", err, tt.errContains)
				}
			} else if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("CoundProvide() error = %v, errContains %v", err, tt.errContains)
			}
		})
	}
}

func Test_coffin_AddToDefault(t *testing.T) {
	type args struct {
		t reflect.Type
	}
	tests := []struct {
		name    string
		coffin  *coffin
		args    args
		wantErr bool
	}{
		{
			name:   "add type to default which cannot be provided",
			coffin: newCoffin(&g1{}),
			args: args{
				t: reflect.TypeOf(g1{}),
			},
			wantErr: true,
		},
		{
			name:   "add type to default",
			coffin: newCoffin(&g1{}),
			args: args{
				t: reflect.TypeOf(new(i1)).Elem(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.coffin
			if err := c.AddToDefault(tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("AddToDefault() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type g1Provider struct {
	err error
	g1  *g1
}

func (g *g1Provider) Provide(tagConf string) (*g1, error) {
	return g.g1, g.err
}

func Test_coffin_Provide(t *testing.T) {
	var x struct {
		g *g1
		i i1
	}

	of := reflect.TypeOf(&x).Elem()
	gField, _ := of.FieldByName("g")
	iField, _ := of.FieldByName("i")
	var g11 = g1{}

	c := newCoffin(&testNamedProvider{})

	c.defaultTypeMap[reflect.TypeOf(&g1{})] = true
	c.defaultTypeMap[reflect.TypeOf(new(i1)).Elem()] = true

	type args struct {
		byName  bool
		tagConf string
		t       reflect.Type
	}
	tests := []struct {
		name    string
		coffin  *coffin
		args    args
		want    any
		wantErr bool
	}{
		{
			name:   "compatible goner for pointer",
			coffin: newCoffin(&g11),
			args: args{
				byName:  false,
				tagConf: "",
				t:       gField.Type,
			},
			want:    &g11,
			wantErr: false,
		},
		{
			name:   "compatible goner for interface",
			coffin: newCoffin(&g11),
			args: args{
				byName:  false,
				tagConf: "",
				t:       iField.Type,
			},
			want:    &g11,
			wantErr: false,
		},
		{
			name: "provider provide success",
			coffin: newCoffin(&g1Provider{
				g1:  &g11,
				err: nil,
			}),
			args: args{
				byName:  false,
				tagConf: "",
				t:       gField.Type,
			},
			want:    &g11,
			wantErr: false,
		},
		{
			name: "provider provide err",
			coffin: newCoffin(&g1Provider{
				g1:  nil,
				err: errors.New("err"),
			}),
			args: args{
				byName:  false,
				tagConf: "",
				t:       gField.Type,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "named provider with default interface type",
			coffin: c,
			args: args{
				t:      iField.Type,
				byName: false,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "named provider with byName = true",
			coffin: newCoffin(&testNamedProvider{}),
			args: args{
				t:      iField.Type,
				byName: true,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "err",
			coffin: newCoffin(&testNamedProvider{}),
			args: args{
				t:      reflect.TypeOf(new(i1)).Elem(),
				byName: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.coffin
			got, err := c.Provide(tt.args.byName, tt.args.tagConf, tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provide() got = %v, want %v", got, tt.want)
			}
		})
	}
}
