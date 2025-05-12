package gone

import (
	"go.uber.org/mock/gomock"
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

func Test_dependenceAnalyzer_analyzerFieldDependencies(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	mockLogger := NewMockLogger(controller)
	mockiKeeper := NewMockiKeeper(controller)

	type inject struct {
		Flag
	}

	var g1 struct {
		Flag
		NormalField string

		GoneField             string `gone:""`
		GoneFieldWithAllowNil string `gone:"" option:"allowNil"`

		inject       `gone:""`
		nameInjected *inject `gone:"name-injected,other-info"`

		injectSlice []*inject `gone:"test-*,xxx,xxx"`
	}

	of := reflect.TypeOf(&g1)
	NormalField, _ := of.Elem().FieldByName("NormalField")
	GoneField, _ := of.Elem().FieldByName("GoneField")
	GoneFieldWithAllowNil, _ := of.Elem().FieldByName("GoneFieldWithAllowNil")
	injectField, _ := of.Elem().FieldByName("inject")
	nameInjectedField, _ := of.Elem().FieldByName("nameInjected")
	injectSliceField, _ := of.Elem().FieldByName("injectSlice")

	var record any

	type args struct {
		field   reflect.StructField
		coName  string
		process func(asSlice bool, extend string, coffins ...*coffin) error
	}
	tests := []struct {
		name    string
		setUp   func() func()
		args    args
		wantErr func(err error) bool
	}{
		{
			name: "do not process normalField which do not tag with gone",
			args: args{
				field:  NormalField,
				coName: "",
				process: func(asSlice bool, extend string, coffins ...*coffin) error {
					t.Fatalf("should not call this")
					return nil
				},
			},
		},
		{
			name: "process normalField which tag with gone",
			setUp: func() func() {
				mockiKeeper.EXPECT().getByTypeAndPattern(GoneField.Type, "*").Return(nil)
				return func() {}
			},
			args: args{
				field:  GoneField,
				coName: "g1",
				process: func(asSlice bool, extend string, coffins ...*coffin) error {
					t.Fatalf("should not call this")
					return nil
				},
			},
			wantErr: func(err error) bool {
				return strings.Contains(err.Error(), "no compatible value found for field \"GoneField\" of \"g1\"")
			},
		},
		{
			name: "process normalField which tag with gone and option allowNil",
			setUp: func() func() {
				mockiKeeper.EXPECT().getByTypeAndPattern(GoneFieldWithAllowNil.Type, "*").Return(nil)
				return func() {}
			},
			args: args{
				field:  GoneFieldWithAllowNil,
				coName: "g1",
				process: func(asSlice bool, extend string, coffins ...*coffin) error {
					t.Fatalf("should not call this")
					return nil
				},
			},
		},
		{
			name: "find multi goner",
			setUp: func() func() {
				mockiKeeper.EXPECT().getByTypeAndPattern(injectField.Type, "*").Return([]*coffin{
					{
						name: "g1",
					},
					{
						name: "g2",
					},
				})
				mockLogger.EXPECT().Warnf("found multiple value without a default when filling filed %q of %q - using first one. ", "inject", "g1")
				return func() {
					if record != true {
						t.Fatalf("process do not exectued")
					}
					record = nil
				}
			},
			args: args{
				field:  injectField,
				coName: "g1",
				process: func(asSlice bool, extend string, coffins ...*coffin) error {
					record = true
					if len(coffins) != 1 {
						t.Fatalf("should find 2 goner")
					}
					if coffins[0].name != "g1" {
						t.Fatalf("should find g1")
					}
					return nil
				},
			},
		},
		{
			name: "find multi goner with default",
			setUp: func() func() {
				mockiKeeper.EXPECT().getByTypeAndPattern(injectField.Type, "*").Return([]*coffin{
					{
						name: "g1",
					},
					{
						name: "g2",
						defaultTypeMap: map[reflect.Type]bool{
							injectField.Type: true,
						},
					},
				})
				return func() {
					if record != true {
						t.Fatalf("process do not exectued")
					}
					record = nil
				}
			},
			args: args{
				field:  injectField,
				coName: "g1",
				process: func(asSlice bool, extend string, coffins ...*coffin) error {
					record = true
					if len(coffins) != 1 {
						t.Fatalf("should find 2 goner")
					}
					if coffins[0].name != "g2" {
						t.Fatalf("should find g1")
					}
					return nil
				},
			},
		},
		{
			name: "find goner by name",
			args: args{
				field:  nameInjectedField,
				coName: "g1",
				process: func(asSlice bool, extend string, coffins ...*coffin) error {
					record = true
					if len(coffins) != 1 {
						t.Fatalf("should find 2 goner")
					}
					return nil
				},
			},
			setUp: func() func() {
				mockiKeeper.EXPECT().getByName("name-injected").Return(&coffin{})

				return func() {
					if record != true {
						t.Fatalf("process do not exectued")
					}
					record = nil
				}
			},
		},
		{
			name: "filed is slice & find slice type",
			args: args{
				field:  injectSliceField,
				coName: "g1",
				process: func(asSlice bool, extend string, coffins ...*coffin) error {
					record = true
					if len(coffins) != 1 {
						t.Fatalf("should be only one goner")
					}
					if asSlice {
						t.Fatalf("should not process as slice")
					}
					if extend != "xxx,xxx" {
						t.Fatalf("extend is not right")
					}
					return nil
				},
			},
			setUp: func() func() {
				mockiKeeper.EXPECT().getByTypeAndPattern(injectSliceField.Type, "test-*").Return([]*coffin{
					{
						name: "test-1",
					},
				})
				return func() {
					if record != true {
						t.Fatalf("process do not exectued")
					}
					record = nil
				}
			},
		},
		{
			name: "filed is slice & find slice element type",
			args: args{
				field:  injectSliceField,
				coName: "g1",
				process: func(asSlice bool, extend string, coffins ...*coffin) error {
					record = true
					if len(coffins) != 2 {
						t.Fatalf("coffins len should be 2")
					}
					if !asSlice {
						t.Fatalf("should process as slice")
					}
					if extend != "xxx,xxx" {
						t.Fatalf("extend is not right")
					}
					return nil
				},
			},
			setUp: func() func() {
				mockiKeeper.EXPECT().getByTypeAndPattern(injectSliceField.Type, "test-*").Return(nil)
				mockiKeeper.EXPECT().getByTypeAndPattern(injectSliceField.Type.Elem(), "test-*").Return([]*coffin{
					{
						name: "test-1",
					},
					{
						name: "test-2",
					},
				})
				return func() {
					if record != true {
						t.Fatalf("process do not exectued")
					}
					record = nil
				}
			},
		},
		{
			name: "filed is slice & find slice element type",
			args: args{
				field:  injectSliceField,
				coName: "g1",
				process: func(asSlice bool, extend string, coffins ...*coffin) error {
					record = true
					return nil
				},
			},
			setUp: func() func() {
				mockiKeeper.EXPECT().getByTypeAndPattern(injectSliceField.Type, "test-*").Return(nil)
				mockiKeeper.EXPECT().getByTypeAndPattern(injectSliceField.Type.Elem(), "test-*").Return(nil)
				return func() {
					if record != nil {
						t.Fatalf("process should not be exectued")
					}
					record = nil
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setUp != nil {
				defer tt.setUp()()
			}

			s := &dependenceAnalyzer{
				iKeeper: mockiKeeper,
				logger:  mockLogger,
			}
			err := s.analyzerFieldDependencies(tt.args.field, tt.args.coName, tt.args.process)
			if tt.wantErr != nil {
				if !tt.wantErr(err) {
					t.Errorf("analyzerFieldDependencies() error = %v, wantErr process failed", err)
				}
			} else if err != nil {
				t.Errorf("analyzerFieldDependencies() error = %v", err)
			}
		})
	}
}
