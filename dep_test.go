package gone

import (
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
