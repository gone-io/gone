package gone

import (
	"reflect"
	"testing"
)

// TestCore_GetDepByName tests the getDepByName method
func TestCore_GetDepByName(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Core)
		depName  string
		wantNil  bool
		wantType reflect.Type
	}{
		{
			name: "Find existing named component",
			setup: func(core *Core) {
				_ = core.Load(&MockNamed{})
			},
			depName:  "named-dep",
			wantNil:  false,
			wantType: reflect.TypeOf(&MockNamed{}),
		},
		{
			name: "Find default provider",
			setup: func(core *Core) {
				// Core itself is loaded as default provider
			},
			depName:  DefaultProviderName,
			wantNil:  false,
			wantType: reflect.TypeOf(&Core{}),
		},
		{
			name: "Component not found",
			setup: func(core *Core) {
				_ = core.Load(&MockComponent{})
			},
			depName: "non-existent",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			coffin, err := core.getDepByName(tt.depName)
			if tt.wantNil {
				if err == nil {
					t.Errorf("getDepByName(%q) expected error, got nil", tt.depName)
				}
				if coffin != nil {
					t.Errorf("getDepByName(%q) expected nil coffin, got %v", tt.depName, coffin)
				}
			} else {
				if err != nil {
					t.Errorf("getDepByName(%q) unexpected error: %v", tt.depName, err)
				}
				if coffin == nil {
					t.Errorf("getDepByName(%q) unexpected nil coffin", tt.depName)
				} else if reflect.TypeOf(coffin.goner) != tt.wantType {
					t.Errorf("getDepByName(%q) got type %v, want %v", tt.depName,
						reflect.TypeOf(coffin.goner), tt.wantType)
				}
			}
		})
	}
}

// TestCore_GetDepByType tests the getDepByType method
func TestCore_GetDepByType(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Core)
		targetType reflect.Type
		wantNil    bool
		wantType   reflect.Type
	}{
		{
			name: "Find existing type",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
			},
			targetType: reflect.TypeOf(&MockDependency{}),
			wantNil:    false,
			wantType:   reflect.TypeOf(&MockDependency{}),
		},
		{
			name: "Multiple implementations with default",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{}, Name("dep1"))
				_ = core.Load(&MockDependency{}, IsDefault(), Name("dep2"))
			},
			targetType: reflect.TypeOf(&MockDependency{}),
			wantNil:    false,
			wantType:   reflect.TypeOf(&MockDependency{}),
		},
		{
			name: "Interface type with implementation",
			setup: func(core *Core) {
				_ = core.Load(&MockInitiator{})
			},
			targetType: reflect.TypeOf((*Initiator)(nil)).Elem(),
			wantNil:    false,
			wantType:   reflect.TypeOf(&MockInitiator{}),
		},
		{
			name:       "Type not found",
			setup:      func(core *Core) {},
			targetType: reflect.TypeOf(&MockComponent{}),
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			coffin, err := core.getDepByType(tt.targetType)
			if tt.wantNil {
				if err == nil {
					t.Errorf("getDepByType(%v) expected error, got nil", tt.targetType)
				}
				if coffin != nil {
					t.Errorf("getDepByType(%v) expected nil coffin, got %v", tt.targetType, coffin)
				}
			} else {
				if err != nil {
					t.Errorf("getDepByType(%v) unexpected error: %v", tt.targetType, err)
				}
				if coffin == nil {
					t.Errorf("getDepByType(%v) unexpected nil coffin", tt.targetType)
				} else if reflect.TypeOf(coffin.goner) != tt.wantType {
					t.Errorf("getDepByType(%v) got type %v, want %v", tt.targetType,
						reflect.TypeOf(coffin.goner), tt.wantType)
				}
			}
		})
	}
}

// TestCore_GetSliceDepsByType tests the getSliceDepsByType method
func TestCore_GetSliceDepsByType(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Core)
		targetType reflect.Type
		wantCount  int
	}{
		{
			name: "No matching components",
			setup: func(core *Core) {
				_ = core.Load(&MockNamed{})
			},
			targetType: reflect.TypeOf([]*MockDependency{}),
			wantCount:  0,
		},
		{
			name: "Single matching component",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
			},
			targetType: reflect.TypeOf([]*MockDependency{}),
			wantCount:  1,
		},
		{
			name: "Multiple matching components",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{}, Name("dep1"))
				_ = core.Load(&MockDependency{}, Name("dep2"))
				_ = core.Load(&MockDependency{}, Name("dep3"))
			},
			targetType: reflect.TypeOf([]*MockDependency{}),
			wantCount:  3,
		},
		{
			name: "Interface type with multiple implementations",
			setup: func(core *Core) {
				_ = core.Load(&MockInitiator{})
				_ = core.Load(&MockInitCombinations{})
			},
			targetType: reflect.TypeOf([]Initiator{}),
			wantCount:  2,
		},
		{
			name: "OnlyForName flag excludes component",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
				_ = core.Load(&MockDependency{}, OnlyForName())
			},
			targetType: reflect.TypeOf([]*MockDependency{}),
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			deps := core.getSliceDepsByType(tt.targetType)
			if len(deps) != tt.wantCount {
				t.Errorf("getSliceDepsByType(%v) returned %d coffins, want %d",
					tt.targetType, len(deps), tt.wantCount)
			}
		})
	}
}
