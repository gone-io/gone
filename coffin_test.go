package gone

import (
	"reflect"
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
