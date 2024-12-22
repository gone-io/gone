package gone

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
			assert.Equal(t, tt.wantMap, gotMap)
			assert.Equal(t, tt.wantKeys, gotKeys)
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
			want: "*github.com/gone-io/gone.LocalType",
		},
		{
			name: "Empty interface",
			t:    reflect.TypeOf(&iface).Elem(),
			want: "interface{}",
		},
		{
			name: "Named struct",
			t:    reflect.TypeOf(LocalType{}),
			want: "github.com/gone-io/gone.LocalType",
		},
		{
			name: "Anonymous struct",
			t:    reflect.TypeOf(struct{}{}),
			want: "struct{}",
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
			want: "github.com/gone-io/gone.TestGetFuncName",
		},
		{
			name: "Anonymous function",
			f:    namedFunc,
			want: "github.com/gone-io/gone.TestGetFuncName.func1",
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
			assert.Equal(t, tt.want, got)
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

	assert.True(t, result.CanAddr(), "BlackMagic result should be addressable")
	assert.Equal(t, 42, result.Interface().(testStruct).Value)
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
	assert.NoError(t, err)
	assert.Equal(t, 1, counter, "Function should be called once")

	// Second call should not execute
	err = wrappedFn(testLoader)
	assert.NoError(t, err)
	assert.Equal(t, 2, counter, "Function should still be called once")
}

func TestSafeExecute(t *testing.T) {
	tests := []struct {
		name      string
		fn        func()
		wantError bool
	}{
		{
			name: "Normal execution",
			fn: func() {
				// Do nothing
			},
			wantError: false,
		},
		{
			name: "Panic execution",
			fn: func() {
				panic("test panic")
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SafeExecute(tt.fn)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
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
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetInterfaceType(t *testing.T) {
	type TestInterface interface {
		Test()
	}

	var ti TestInterface
	got := GetInterfaceType(&ti)

	assert.Equal(t, reflect.Interface, got.Kind())
	assert.Equal(t, "TestInterface", got.Name())
}

// Mock Loader for testing OnceLoad
type mockLoader struct {
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

	assert.NotEqual(t, key1, key2, "Generated keys should be unique")
}
