package gone

import (
	"reflect"
	"testing"
)

// TestCore_GetCoffinsByType tests the getCoffinsByType method
func TestCore_GetCoffinsByType(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Core)
		targetType reflect.Type
		wantCount  int
	}{
		{
			name: "No matching coffins",
			setup: func(core *Core) {
				// Load components that don't match the target type
				_ = core.Load(&MockNamed{})
			},
			targetType: reflect.TypeOf(&MockDependency{}),
			wantCount:  0,
		},
		{
			name: "Single matching coffin",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
			},
			targetType: reflect.TypeOf(&MockDependency{}),
			wantCount:  1,
		},
		{
			name: "Multiple matching coffins",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
				_ = core.Load(&MockDependency{})
				_ = core.Load(&MockDependency{}, Name("named-dep"))
			},
			targetType: reflect.TypeOf(&MockDependency{}),
			wantCount:  3,
		},
		{
			name: "OnlyForName flag excludes coffin",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
				_ = core.Load(&MockDependency{}, OnlyForName())
			},
			targetType: reflect.TypeOf(&MockDependency{}),
			wantCount:  1,
		},
		{
			name: "Interface type matching",
			setup: func(core *Core) {
				_ = core.Load(&MockInitiator{})
				_ = core.Load(&MockInitCombinations{})
			},
			targetType: reflect.TypeOf((*Initiator)(nil)).Elem(),
			wantCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			coffins := core.getCoffinsByType(tt.targetType)
			if len(coffins) != tt.wantCount {
				t.Errorf("getCoffinsByType() returned %d coffins, want %d", len(coffins), tt.wantCount)
			}
		})
	}
}

// TestCore_GetDefaultCoffinByType tests the getDefaultCoffinByType method
func TestCore_GetDefaultCoffinByType(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*Core)
		targetType  reflect.Type
		wantNil     bool
		wantDefault bool
	}{
		{
			name: "No matching coffins",
			setup: func(core *Core) {
				// Load components that don't match the target type
				_ = core.Load(&MockNamed{})
			},
			targetType:  reflect.TypeOf(&MockDependency{}),
			wantNil:     true,
			wantDefault: false,
		},
		{
			name: "Single matching coffin",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
			},
			targetType:  reflect.TypeOf(&MockDependency{}),
			wantNil:     false,
			wantDefault: false,
		},
		{
			name: "Multiple coffins with default",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{})
				_ = core.Load(&MockDependency{}, IsDefault())
			},
			targetType:  reflect.TypeOf(&MockDependency{}),
			wantNil:     false,
			wantDefault: true,
		},
		{
			name: "Multiple coffins without default",
			setup: func(core *Core) {
				_ = core.Load(&MockDependency{}, Name("dep1"))
				_ = core.Load(&MockDependency{}, Name("dep2"))
			},
			targetType:  reflect.TypeOf(&MockDependency{}),
			wantNil:     false,
			wantDefault: false,
		},
		{
			name: "Interface type with default implementation",
			setup: func(core *Core) {
				_ = core.Load(&MockInitiator{})
				_ = core.Load(&MockInitCombinations{}, IsDefault((*Initiator)(nil)))
			},
			targetType:  reflect.TypeOf((*Initiator)(nil)).Elem(),
			wantNil:     false,
			wantDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := NewCore()
			tt.setup(core)

			coffin := core.getDefaultCoffinByType(tt.targetType)
			if (coffin == nil) != tt.wantNil {
				t.Errorf("getDefaultCoffinByType() returned nil: %v, want nil: %v", coffin == nil, tt.wantNil)
			}

			if !tt.wantNil && tt.wantDefault {
				if !coffin.isDefault(tt.targetType) {
					t.Errorf("getDefaultCoffinByType() did not return the default coffin")
				}
			}
		})
	}
}
