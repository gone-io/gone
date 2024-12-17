package gone

import (
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

			if !reflect.DeepEqual(gotMap, tt.wantMap) {
				t.Errorf("TagStringParse() map = %v, want %v", gotMap, tt.wantMap)
			}
			if !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("TagStringParse() keys = %v, want %v", gotKeys, tt.wantKeys)
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
			name:       "Name only",
			tag:        "myGoner",
			wantName:   "myGoner",
			wantExtend: "",
		},
		{
			name:       "Name and extend",
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

type testStruct struct {
	name string
}

func (t *testStruct) TestMethod() {}

func TestIsCompatible(t *testing.T) {
	var ti testInterface
	ts := &testStruct{}

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
			t:     reflect.TypeOf(&testStruct{}),
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
			t:     reflect.TypeOf(&testStruct{}),
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
	a := &testStruct{
		name: "a",
	}
	b := &testStruct{
		name: "b",
	}
	c := &testStruct{
		name: "c",
	}

	tests := []struct {
		name string
		list []*testStruct
		want []*testStruct
	}{
		{
			name: "No duplicates",
			list: []*testStruct{a, b, c},
			want: []*testStruct{a, b, c},
		},
		{
			name: "With duplicates",
			list: []*testStruct{a, b, a, c, b},
			want: []*testStruct{a, b, c},
		},
		{
			name: "Empty list",
			list: []*testStruct{},
			want: []*testStruct{},
		},
		{
			name: "Single element",
			list: []*testStruct{a},
			want: []*testStruct{a},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveRepeat(tt.list)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveRepeat() = %v, want %v", got, tt.want)
			}
		})
	}
}
