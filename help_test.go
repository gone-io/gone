package gone

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestTagStringParse(t *testing.T) {
	tests := []struct {
		name     string
		conf     string
		wantMap  map[string]string
		wantKeys []string
	}{
		{
			name: "Simple key-value pairs",
			conf: "key1=value1,key2=value2",
			wantMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantKeys: []string{"key1", "key2"},
		},
		{
			name: "Empty values",
			conf: "key1=,key2=value2",
			wantMap: map[string]string{
				"key1": "",
				"key2": "value2",
			},
			wantKeys: []string{"key1", "key2"},
		},
		{
			name: "No values",
			conf: "key1,key2",
			wantMap: map[string]string{
				"key1": "",
				"key2": "",
			},
			wantKeys: []string{"key1", "key2"},
		},
		{
			name: "With spaces",
			conf: " key1 = value1 , key2 = value2 ",
			wantMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantKeys: []string{"key1", "key2"},
		},
		{
			name:     "Empty string",
			conf:     "",
			wantMap:  map[string]string{"": ""},
			wantKeys: []string{""},
		},
		{
			name: "Duplicate keys",
			conf: "key1=value1,key1=value2",
			wantMap: map[string]string{
				"key1": "value2",
			},
			wantKeys: []string{"key1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMap, gotKeys := TagStringParse(tt.conf)
			if !reflect.DeepEqual(tt.wantMap, gotMap) {
				t.Errorf("TagStringParse() gotMap = %v, want %v", gotMap, tt.wantMap)
			}
			if !reflect.DeepEqual(tt.wantKeys, gotKeys) {
				t.Errorf("TagStringParse() gotKeys = %v, want %v", gotKeys, tt.wantKeys)
			}
		})
	}
}

func TestParseGoneTag(t *testing.T) {
	tests := []struct {
		name       string
		tag        string
		wantName   string
		wantExtend string
	}{
		{
			name:       "GonerName only",
			tag:        "myGoner",
			wantName:   "myGoner",
			wantExtend: "",
		},
		{
			name:       "GonerName and extend",
			tag:        "myGoner,config=value",
			wantName:   "myGoner",
			wantExtend: "config=value",
		},
		{
			name:       "Empty string",
			tag:        "",
			wantName:   "",
			wantExtend: "",
		},
		{
			name:       "Multiple commas",
			tag:        "myGoner,key1=value1,key2=value2",
			wantName:   "myGoner",
			wantExtend: "key1=value1,key2=value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotExtend := ParseGoneTag(tt.tag)
			if gotName != tt.wantName {
				t.Errorf("ParseGoneTag() name = %v, want %v", gotName, tt.wantName)
			}
			if gotExtend != tt.wantExtend {
				t.Errorf("ParseGoneTag() extend = %v, want %v", gotExtend, tt.wantExtend)
			}
		})
	}
}

type testInterface interface {
	TestMethod()
}

type hTestStruct struct {
	name string
}

func (t *hTestStruct) TestMethod() {}

func TestIsCompatible(t *testing.T) {
	var ti testInterface
	ts := &hTestStruct{}

	tests := []struct {
		name  string
		t     reflect.Type
		goner any
		want  bool
	}{
		{
			name:  "Interface implementation",
			t:     reflect.TypeOf(&ti).Elem(),
			goner: ts,
			want:  true,
		},
		{
			name:  "Exact type match",
			t:     reflect.TypeOf(&hTestStruct{}),
			goner: ts,
			want:  true,
		},
		{
			name:  "Type mismatch",
			t:     reflect.TypeOf(""),
			goner: ts,
			want:  false,
		},
		{
			name:  "Nil goner",
			t:     reflect.TypeOf(&hTestStruct{}),
			goner: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCompatible(tt.t, tt.goner); got != tt.want {
				t.Errorf("IsCompatible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTypeName(t *testing.T) {
	type LocalType struct{}

	var ch = make(chan Flag)
	type Ch chan Flag

	var iface interface{}

	tests := []struct {
		name string
		t    reflect.Type
		want string
	}{
		{
			name: "Basic type",
			t:    reflect.TypeOf(""),
			want: "string",
		},
		{
			name: "Array",
			t:    reflect.TypeOf([3]int{}),
			want: "[3]int",
		},
		{
			name: "Slice",
			t:    reflect.TypeOf([]string{}),
			want: "[]string",
		},
		{
			name: "Map",
			t:    reflect.TypeOf(map[string]int{}),
			want: "map[string]int",
		},
		{
			name: "Pointer",
			t:    reflect.TypeOf(&LocalType{}),
			want: "*github.com/gone-io/gone/v2.LocalType",
		},
		{
			name: "Empty interface",
			t:    reflect.TypeOf(&iface).Elem(),
			want: "interface{}",
		},
		{
			name: "Named struct",
			t:    reflect.TypeOf(LocalType{}),
			want: "github.com/gone-io/gone/v2.LocalType",
		},
		{
			name: "Anonymous struct",
			t:    reflect.TypeOf(struct{}{}),
			want: "struct{}",
		},
		{
			name: "chan",
			t:    reflect.TypeOf(ch),
			want: "chan gone.Flag",
		},
		{
			name: "Named chan",
			t:    reflect.TypeOf(Ch(ch)),
			want: "github.com/gone-io/gone/v2.Ch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTypeName(tt.t); got != tt.want {
				t.Errorf("GetTypeName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFuncName(t *testing.T) {
	namedFunc := func() {}

	tests := []struct {
		name string
		f    any
		want string
	}{
		{
			name: "Named function",
			f:    TestGetFuncName,
			want: "github.com/gone-io/gone/v2.TestGetFuncName",
		},
		{
			name: "Anonymous function",
			f:    namedFunc,
			want: "github.com/gone-io/gone/v2.TestGetFuncName.func1",
		},
		{
			name: "Non-function",
			f:    "not a function",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFuncName(tt.f); !strings.HasSuffix(got, tt.want) {
				t.Errorf("GetFuncName() = %v, want suffix %v", got, tt.want)
			}
		})
	}
}

func TestRemoveRepeat(t *testing.T) {
	a := &hTestStruct{
		name: "a",
	}
	b := &hTestStruct{
		name: "b",
	}
	c := &hTestStruct{
		name: "c",
	}

	tests := []struct {
		name string
		list []*hTestStruct
		want []*hTestStruct
	}{
		{
			name: "No duplicates",
			list: []*hTestStruct{a, b, c},
			want: []*hTestStruct{a, b, c},
		},
		{
			name: "With duplicates",
			list: []*hTestStruct{a, b, a, c, b},
			want: []*hTestStruct{a, b, c},
		},
		{
			name: "Empty list",
			list: []*hTestStruct{},
			want: []*hTestStruct{},
		},
		{
			name: "Single element",
			list: []*hTestStruct{a},
			want: []*hTestStruct{a},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveRepeat(tt.list)
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("RemoveRepeat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlackMagic(t *testing.T) {
	type testStruct struct {
		Value int
	}
	v := testStruct{Value: 42}
	rv := reflect.ValueOf(&v).Elem()

	result := BlackMagic(rv)

	if !result.CanAddr() {
		t.Error("BlackMagic result should be addressable")
	}
	if result.Interface().(testStruct).Value != 42 {
		t.Errorf("BlackMagic() value = %v, want %v", result.Interface().(testStruct).Value, 42)
	}
}

func TestOnceLoad(t *testing.T) {
	counter := 0
	testLoader := &mockLoader{}

	fn := func(loader Loader) error {
		counter++
		return nil
	}

	wrappedFn := OnceLoad(fn)

	// First call should execute
	err := wrappedFn(testLoader)
	if err != nil {
		t.Errorf("OnceLoad() first call error = %v", err)
	}
	if counter != 1 {
		t.Errorf("Function should be called once, got called %d times", counter)
	}

	// Second call should not execute
	err = wrappedFn(testLoader)
	if err != nil {
		t.Errorf("OnceLoad() second call error = %v", err)
	}
	if counter != 2 {
		t.Errorf("Function should be called twice, got called %d times", counter)
	}
}

func TestSafeExecute(t *testing.T) {
	tests := []struct {
		name      string
		fn        func() error
		wantError bool
	}{
		{
			name: "Normal execution",
			fn: func() error {
				// Do nothing
				return nil
			},
			wantError: false,
		},
		{
			name: "Panic execution",
			fn: func() error {
				panic("test panic")
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SafeExecute(tt.fn)
			if tt.wantError && err == nil {
				t.Error("SafeExecute() expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("SafeExecute() unexpected error = %v", err)
			}
		})
	}
}

func TestConvertUppercaseCamel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple string",
			input: "hello",
			want:  "HELLO",
		},
		{
			name:  "Dotted string",
			input: "hello.world",
			want:  "HELLO_WORLD",
		},
		{
			name:  "Multiple dots",
			input: "test.hello.world",
			want:  "TEST_HELLO_WORLD",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertUppercaseCamel(tt.input)
			if got != tt.want {
				t.Errorf("convertUppercaseCamel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetInterfaceType(t *testing.T) {
	type TestInterface interface {
		Test()
	}

	var ti TestInterface
	got := GetInterfaceType(&ti)

	if got.Kind() != reflect.Interface {
		t.Errorf("GetInterfaceType() kind = %v, want %v", got.Kind(), reflect.Interface)
	}
	if got.Name() != "TestInterface" {
		t.Errorf("GetInterfaceType() name = %v, want %v", got.Name(), "TestInterface")
	}
}

// Mock Loader for testing OnceLoad
type mockLoader struct {
	Loader
	loadedKeys map[LoaderKey]bool
}

func (m *mockLoader) Load(goner Goner, options ...Option) error {
	return nil
}

func (m *mockLoader) Loaded(key LoaderKey) bool {
	if m.loadedKeys == nil {
		m.loadedKeys = make(map[LoaderKey]bool)
	}
	return m.loadedKeys[key]
}

func TestGenLoaderKey(t *testing.T) {
	key1 := GenLoaderKey()
	key2 := GenLoaderKey()

	if key1 == key2 {
		t.Error("Generated keys should be unique")
	}
}

func TestLoadFunc(t *testing.T) {
	type TestGoner struct {
		Flag
		i int
	}
	var goner TestGoner

	t.Run("case 1: use OnceLoad once outside", func(t *testing.T) {
		originLoader := func(loader Loader) error {
			return loader.Load(&goner)
		}

		loadFunc := originLoader

		NewApp(loadFunc, loadFunc).
			Run(func(goners []*TestGoner) {
				if len(goners) != 1 {
					t.Errorf("Expected 1 goner, got %d", len(goners))
				}
			})
	})

	t.Run("case 2: use OnceLoad inside more than once", func(t *testing.T) {
		loadFunc := func(loader Loader) error {
			return OnceLoad(func(loader Loader) error {
				return loader.Load(&goner)
			})(loader)
		}

		NewApp(loadFunc, loadFunc).
			Run(func(goners []*TestGoner) {
				if len(goners) != 1 {
					t.Errorf("Expected 1 goner, got %d", len(goners))
				}
			})
	})
	t.Run("case 3: use OnceLoad inside more wrapped func dep on `Closure`", func(t *testing.T) {
		loadFunc := func(loader Loader) error {
			i := 100
			return OnceLoad(func(loader Loader) error {
				return loader.Load(&TestGoner{i: i})
			})(loader)
		}

		NewApp(loadFunc, loadFunc).
			Run(func(goners []*TestGoner) {
				if len(goners) != 1 {
					t.Errorf("Expected 1 goner, got %d", len(goners))
				}
			})
	})
}

func TestBuildThirdComponentLoadFunc(t *testing.T) {
	type TestComponent struct {
		i int
	}
	var goner TestComponent

	loadFunc := func(loader Loader) error {
		return BuildThirdComponentLoadFunc(&goner)(loader)
	}

	NewApp(loadFunc, loadFunc).
		Run(func(goners []*TestComponent) {
			if len(goners) != 1 {
				t.Errorf("Expected 1 goner, got %d", len(goners))
			}
		})
}

func TestX(t *testing.T) {
	type TestComponent struct {
		i int
	}
	type TestComponent2 struct {
		i int
	}
	loadFunc1 := BuildThirdComponentLoadFunc[*TestComponent](&TestComponent{i: 10})
	loadFunc2 := BuildThirdComponentLoadFunc[*TestComponent2](&TestComponent2{i: 20})

	key1 := fmt.Sprintf("loadFunc1=> %#v\n", loadFunc1)
	key2 := fmt.Sprintf("loadFunc2=> %#v\n", loadFunc2)
	if key1 == key2 {
		t.Errorf("loadFunc1 should not equal to loadFunc2")
	}
}

func TestIsError(t *testing.T) {
	type args struct {
		err  error
		code int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is error",
			args: args{
				err:  NewError(1, "test", 500),
				code: 500,
			},
		},
		{
			name: "is not error",
			args: args{
				err:  NewError(2, "test", 500),
				code: 1,
			},
			want: false,
		},
		{
			name: "is not error",
			args: args{
				err:  errors.New("test"),
				code: 1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsError(tt.args.err, tt.args.code); got != tt.want {
				t.Errorf("IsError() = %v, want %v", got, tt.want)
			}
		})
	}
}
