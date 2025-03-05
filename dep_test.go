package gone

import (
	"reflect"
	"strings"
	"testing"
)

// createTestCoffin is a helper function to create test coffins
func createTestCoffin(name string) *coffin {
	return &coffin{
		name:  name,
		goner: struct{}{},
	}
}

func TestCheckCircularDepsAndGetBestInitOrder(t *testing.T) {
	tests := []struct {
		name             string
		initiatorDepsMap map[*coffin][]*coffin
		wantCircular     bool
		wantOrderLen     int
	}{
		{
			name: "Linear dependency chain",
			initiatorDepsMap: func() map[*coffin][]*coffin {
				a := createTestCoffin("A")
				b := createTestCoffin("B")
				c := createTestCoffin("C")
				return map[*coffin][]*coffin{
					a: {b},
					b: {c},
					c: {},
				}
			}(),
			wantCircular: false,
			wantOrderLen: 3,
		},
		{
			name: "Circular dependency",
			initiatorDepsMap: func() map[*coffin][]*coffin {
				a := createTestCoffin("A")
				b := createTestCoffin("B")
				c := createTestCoffin("C")
				return map[*coffin][]*coffin{
					a: {b},
					b: {c},
					c: {a}, // Creates a cycle
				}
			}(),
			wantCircular: true,
			wantOrderLen: 0,
		},
		{
			name:             "Empty dependency map",
			initiatorDepsMap: map[*coffin][]*coffin{},
			wantCircular:     false,
			wantOrderLen:     0,
		},
		{
			name: "Diamond dependency",
			initiatorDepsMap: func() map[*coffin][]*coffin {
				a := createTestCoffin("A")
				b := createTestCoffin("B")
				c := createTestCoffin("C")
				d := createTestCoffin("D")
				return map[*coffin][]*coffin{
					a: {b, c},
					b: {d},
					c: {d},
					d: {},
				}
			}(),
			wantCircular: false,
			wantOrderLen: 4,
		},
		{
			name: "Self dependency",
			initiatorDepsMap: func() map[*coffin][]*coffin {
				a := createTestCoffin("A")
				return map[*coffin][]*coffin{
					a: {a}, // Self dependency
				}
			}(),
			wantCircular: true,
			wantOrderLen: 0,
		},
		{
			name: "Multiple circular dependencies",
			initiatorDepsMap: func() map[*coffin][]*coffin {
				a := createTestCoffin("A")
				b := createTestCoffin("B")
				c := createTestCoffin("C")
				d := createTestCoffin("D")
				return map[*coffin][]*coffin{
					a: {b},
					b: {c},
					c: {a}, // First cycle
					d: {d}, // Second cycle (self dependency)
				}
			}(),
			wantCircular: true,
			wantOrderLen: 0,
		},
		{
			name: "Single node with no dependencies",
			initiatorDepsMap: func() map[*coffin][]*coffin {
				a := createTestCoffin("A")
				return map[*coffin][]*coffin{
					a: {},
				}
			}(),
			wantCircular: false,
			wantOrderLen: 1,
		},
		{
			name: "Node with empty dependency slice",
			initiatorDepsMap: func() map[*coffin][]*coffin {
				a := createTestCoffin("A")
				b := createTestCoffin("B")
				return map[*coffin][]*coffin{
					a: {b},
					b: make([]*coffin, 0), // Explicitly empty slice
				}
			}(),
			wantCircular: false,
			wantOrderLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			circularDeps, initOrder := checkCircularDepsAndGetBestInitOrder(tt.initiatorDepsMap)

			// Check circular dependency detection
			if (len(circularDeps) > 0) != tt.wantCircular {
				t.Errorf("checkCircularDepsAndGetBestInitOrder() circular = %v, want %v",
					len(circularDeps) > 0, tt.wantCircular)
			}

			// Check initialization order length
			if len(initOrder) != tt.wantOrderLen {
				t.Errorf("checkCircularDepsAndGetBestInitOrder() order length = %v, want %v",
					len(initOrder), tt.wantOrderLen)
			}

			if len(initOrder) > 0 {
				// Verify the initialization order is valid
				seen := make(map[*coffin]bool)
				for _, co := range initOrder {
					// Check that all dependencies of current coffin have been initialized
					for _, dep := range tt.initiatorDepsMap[co] {
						if !seen[dep] {
							t.Errorf("Invalid initialization order: %v depends on %v but it's not initialized yet",
								co.name, dep.name)
						}
					}
					seen[co] = true
				}
			}
		})
	}
}

type testStruct struct {
	Dep1 *string `gone:"dep1"`
	Dep2 *int    `gone:"dep2"`
}

type testFunc func(*string, *int)

func TestGetGonerFillDeps(t *testing.T) {
	tests := []struct {
		name          string
		goner         interface{}
		setupCore     func(*Core)
		wantDepsCount int
		wantErr       bool
	}{
		{
			name:  "Struct with dependencies",
			goner: &testStruct{},
			setupCore: func(c *Core) {
				dep1 := &coffin{name: "dep1", needInitBeforeUse: true}
				dep2 := &coffin{name: "dep2", needInitBeforeUse: true}
				c.nameMap = map[string]*coffin{
					"dep1": dep1,
					"dep2": dep2,
				}
			},
			wantDepsCount: 2,
			wantErr:       false,
		},
		{
			name:  "Struct with missing dependency",
			goner: &testStruct{},
			setupCore: func(c *Core) {
				dep1 := &coffin{name: "dep1", needInitBeforeUse: true}
				c.nameMap = map[string]*coffin{
					"dep1": dep1,
					// dep2 is missing
				}
			},
			wantDepsCount: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := &Core{}
			tt.setupCore(core)

			co := &coffin{
				goner: tt.goner,
			}

			deps, err := core.getGonerFillDeps(co)

			if (err != nil) != tt.wantErr {
				t.Errorf("getGonerFillDeps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(deps) != tt.wantDepsCount {
				t.Errorf("getGonerFillDeps() got %d dependencies, want %d", len(deps), tt.wantDepsCount)
			}
		})
	}
}

func TestGetGonerDeps(t *testing.T) {
	tests := []struct {
		name              string
		coffin            *coffin
		setupCore         func(*Core)
		wantFillDepsCount int
		wantInitDepsCount int
		wantErr           bool
	}{
		{
			name: "Normal coffin with dependencies",
			coffin: &coffin{
				goner:    &testStruct{},
				lazyFill: false,
			},
			setupCore: func(c *Core) {
				dep1 := &coffin{name: "dep1", needInitBeforeUse: true}
				dep2 := &coffin{name: "dep2", needInitBeforeUse: true}
				c.nameMap = map[string]*coffin{
					"dep1": dep1,
					"dep2": dep2,
				}
			},
			wantFillDepsCount: 2,
			wantInitDepsCount: 1, // The coffin itself needs initialization
			wantErr:           false,
		},
		{
			name: "Lazy fill coffin",
			coffin: &coffin{
				goner:    &testStruct{},
				lazyFill: true,
			},
			setupCore: func(c *Core) {
				dep1 := &coffin{name: "dep1", needInitBeforeUse: true}
				dep2 := &coffin{name: "dep2", needInitBeforeUse: true}
				c.nameMap = map[string]*coffin{
					"dep1": dep1,
					"dep2": dep2,
				}
			},
			wantFillDepsCount: 2,
			wantInitDepsCount: 0, // No init dependencies due to lazy fill
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := &Core{}
			tt.setupCore(core)

			fillDeps, initDeps, err := core.getGonerDeps(tt.coffin)

			if (err != nil) != tt.wantErr {
				t.Errorf("getGonerDeps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(fillDeps) != tt.wantFillDepsCount {
					t.Errorf("getGonerDeps() got %d fill dependencies, want %d", len(fillDeps), tt.wantFillDepsCount)
				}
				if len(initDeps) != tt.wantInitDepsCount {
					t.Errorf("getGonerDeps() got %d init dependencies, want %d", len(initDeps), tt.wantInitDepsCount)
				}
			}
		})
	}
}

func TestGetDepByName(t *testing.T) {
	tests := []struct {
		name      string
		setupCore func(*Core)
		depName   string
		wantErr   bool
	}{
		{
			name: "Existing dependency",
			setupCore: func(c *Core) {
				c.nameMap = map[string]*coffin{
					"test": {name: "test"},
				}
			},
			depName: "test",
			wantErr: false,
		},
		{
			name: "Non-existing dependency",
			setupCore: func(c *Core) {
				c.nameMap = map[string]*coffin{}
			},
			depName: "missing",
			wantErr: true,
		},
		{
			name: "Empty name",
			setupCore: func(c *Core) {
				c.nameMap = map[string]*coffin{}
			},
			depName: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := &Core{}
			tt.setupCore(core)

			got, err := core.getDepByName(tt.depName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDepByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("getDepByName() returned nil but expected a coffin")
			}
		})
	}
}

func TestGetDepByType(t *testing.T) {
	type testStruct struct{}
	tests := []struct {
		name      string
		setupCore func(*Core)
		typeToGet reflect.Type
		wantErr   bool
		errMsg    string
	}{
		{
			name: "Existing type in default coffins",
			setupCore: func(c *Core) {
				c.coffins = []*coffin{{
					goner: &testStruct{},
				}}
			},
			typeToGet: reflect.TypeOf(&testStruct{}),
			wantErr:   false,
		},
		{
			name: "Existing type in provider map",
			setupCore: func(c *Core) {
				c.typeProviderDepMap = map[reflect.Type]*coffin{
					reflect.TypeOf(&testStruct{}): {name: "provider"},
				}
			},
			typeToGet: reflect.TypeOf(&testStruct{}),
			wantErr:   false,
		},
		{
			name: "Non-pointer struct type",
			setupCore: func(c *Core) {
				c.coffins = []*coffin{}
			},
			typeToGet: reflect.TypeOf(testStruct{}),
			wantErr:   true,
			errMsg:    "Maybe, you should use A Pointer to this type?",
		},
		{
			name: "Non-existing type",
			setupCore: func(c *Core) {
				c.coffins = []*coffin{}
			},
			typeToGet: reflect.TypeOf(&struct{}{}),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := &Core{
				log: GetDefaultLogger(),
			}
			tt.setupCore(core)

			got, err := core.getDepByType(tt.typeToGet)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDepByType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("getDepByType() error message = %v, want to contain %v", err, tt.errMsg)
				}
			}

			if !tt.wantErr && got == nil {
				t.Error("getDepByType() returned nil but expected a coffin")
			}
		})
	}
}
