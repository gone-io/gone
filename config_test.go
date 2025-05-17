package gone

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
)

// SimpleMockConfigure implements Configure interface for testing
type SimpleMockConfigure struct {
	values map[string]string
}

func (m *SimpleMockConfigure) Get(key string, v any, defaultVal string) error {
	if key == "" {
		return errors.New("key is empty")
	}
	val, exists := m.values[key]
	if !exists {
		val = defaultVal
	}

	if reflect.TypeOf(v).Kind() == reflect.Ptr && reflect.TypeOf(v).Elem().Kind() == reflect.String {
		reflect.ValueOf(v).Elem().SetString(val)
		return nil
	}

	return json.Unmarshal([]byte(val), v)
}

func TestConfigProvider_Name(t *testing.T) {
	provider := &ConfigProvider{}
	if got := provider.GonerName(); got != "config" {
		t.Errorf("ConfigProvider.GonerName() = %v, want %v", got, "config")
	}
}

func TestConfigProvider_Provide(t *testing.T) {
	mockConfigure := &SimpleMockConfigure{
		values: map[string]string{
			"test-key":   `test-value`,
			"number-key": "42",
			//"missing-key": "",
			"invalid-key": "invalid-json",
			"complex-key": `{"name":"test","value":123}`,
		},
	}

	provider := &ConfigProvider{
		configure: mockConfigure,
	}

	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name      string
		tagConf   string
		valueType reflect.Type
		want      any
		wantErr   bool
	}{
		{
			name:      "miss key",
			tagConf:   "",
			valueType: reflect.TypeOf(""),
			want:      "",
			wantErr:   true,
		},
		{
			name:      "String value",
			tagConf:   "test-key",
			valueType: reflect.TypeOf(""),
			want:      "test-value",
			wantErr:   false,
		},
		{
			name:      "Number value",
			tagConf:   "number-key",
			valueType: reflect.TypeOf(0),
			want:      42,
			wantErr:   false,
		},
		{
			name:      "Default value",
			tagConf:   "missing-key=default-value",
			valueType: reflect.TypeOf(""),
			want:      "default-value",
			wantErr:   false,
		},
		{
			name:    "Invalid JSON",
			tagConf: "invalid-key",
			valueType: reflect.TypeOf(struct {
				Name string
			}{}),
			want:    nil,
			wantErr: true,
		},
		{
			name:      "Complex struct",
			tagConf:   "complex-key",
			valueType: reflect.TypeOf(testStruct{}),
			want:      testStruct{Name: "test", Value: 123},
			wantErr:   false,
		},
		{
			name:      "Empty key",
			tagConf:   "",
			valueType: reflect.TypeOf(""),
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.Provide(tt.tagConf, tt.valueType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigProvider.Provide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConfigProvider.Provide() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvConfigure_Get(t *testing.T) {
	// Setup test environment variables
	envVars := map[string]string{
		"TEST_STRING":     "test-value",
		"TEST_INT":        "42",
		"TEST_INT64":      "9223372036854775807",
		"TEST_FLOAT":      "3.14",
		"TEST_BOOL":       "true",
		"TEST_UINT":       "123",
		"TEST_UINT_ERROR": "1ss2",
		"TEST_UINT64":     "18446744073709551615",
		"TEST_DURATION":   "1h30m",
		"TEST_STRUCT":     `{"name":"test","value":123}`,
	}

	for k, v := range envVars {
		os.Setenv(GONE+"_"+k, v)
		defer os.Unsetenv(k)
	}

	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name       string
		key        string
		defaultVal string
		value      any
		want       any
		wantErr    bool
	}{
		{
			name:       "String value",
			key:        "TEST_STRING",
			defaultVal: "default",
			value:      new(string),
			want:       "test-value",
			wantErr:    false,
		},
		{
			name:       "Int value",
			key:        "TEST_INT",
			defaultVal: "0",
			value:      new(int),
			want:       42,
			wantErr:    false,
		},
		{
			name:       "Int64 value",
			key:        "TEST_INT64",
			defaultVal: "0",
			value:      new(int64),
			want:       int64(9223372036854775807),
			wantErr:    false,
		},
		{
			name:       "Float64 value",
			key:        "TEST_FLOAT",
			defaultVal: "0.0",
			value:      new(float64),
			want:       3.14,
			wantErr:    false,
		},
		{
			name:       "Bool value",
			key:        "TEST_BOOL",
			defaultVal: "false",
			value:      new(bool),
			want:       true,
			wantErr:    false,
		},
		{
			name:       "Uint value",
			key:        "TEST_UINT",
			defaultVal: "0",
			value:      new(uint),
			want:       uint(123),
			wantErr:    false,
		},
		{
			name:       "int32 value error",
			key:        "TEST_UINT_ERROR",
			defaultVal: "0",
			value:      new(int32),
			want:       int32(123),
			wantErr:    true,
		},
		{
			name:       "uint32 value error",
			key:        "TEST_UINT_ERROR",
			defaultVal: "0",
			value:      new(uint32),
			want:       uint32(123),
			wantErr:    true,
		},
		{
			name:       "int64 value error",
			key:        "TEST_UINT_ERROR",
			defaultVal: "0",
			value:      new(int64),
			want:       int64(123),
			wantErr:    true,
		},
		{
			name:       "uint64 value error",
			key:        "TEST_UINT_ERROR",
			defaultVal: "0",
			value:      new(uint64),
			want:       uint64(123),
			wantErr:    true,
		},
		{
			name:       "float32 value error",
			key:        "TEST_UINT_ERROR",
			defaultVal: "0",
			value:      new(float32),
			want:       float32(123),
			wantErr:    true,
		},
		{
			name:       "Uint64 value",
			key:        "TEST_UINT64",
			defaultVal: "0",
			value:      new(uint64),
			want:       uint64(18446744073709551615),
			wantErr:    false,
		},
		{
			name:       "Duration value",
			key:        "TEST_DURATION",
			defaultVal: "1s",
			value:      new(time.Duration),
			want:       90 * time.Minute,
			wantErr:    false,
		},
		{
			name:       "Struct value",
			key:        "TEST_STRUCT",
			defaultVal: "{}",
			value:      &testStruct{},
			want:       testStruct{Name: "test", Value: 123},
			wantErr:    false,
		},
		{
			name:       "Default value",
			key:        "NON_EXISTENT",
			defaultVal: "default-value",
			value:      new(string),
			want:       "default-value",
			wantErr:    false,
		},
		{
			name:       "Invalid type",
			key:        "TEST_STRING",
			defaultVal: "",
			value:      "not-a-pointer",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "Invalid number format",
			key:        "TEST_STRING",
			defaultVal: "",
			value:      new(int),
			want:       0,
			wantErr:    true,
		},
	}

	configure := &EnvConfigure{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := configure.Get(tt.key, tt.value, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnvConfigure.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got := reflect.ValueOf(tt.value).Elem().Interface()
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("EnvConfigure.Get() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestConfigProvider_Init(t *testing.T) {
	provider := &ConfigProvider{}
	// Init() should not panic
	provider.Init()
}

func TestConfigProvider_ProvideWithDefaultTag(t *testing.T) {
	mockConfigure := &SimpleMockConfigure{
		values: map[string]string{},
	}

	provider := &ConfigProvider{
		configure: mockConfigure,
	}

	tests := []struct {
		name      string
		tagConf   string
		valueType reflect.Type
		want      any
		wantErr   bool
	}{
		{
			name:      "Use default tag",
			tagConf:   "missing-key,default=default-value",
			valueType: reflect.TypeOf(""),
			want:      "default-value",
			wantErr:   false,
		},
		{
			name:      "Multiple keys with default",
			tagConf:   "key1,key2,default=fallback",
			valueType: reflect.TypeOf(""),
			want:      "fallback",
			wantErr:   false,
		},
		{
			name:      "Empty default value",
			tagConf:   "missing-key,default=",
			valueType: reflect.TypeOf(""),
			want:      "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.Provide(tt.tagConf, tt.valueType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigProvider.Provide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConfigProvider.Provide() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetValueByReflectValue(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		strVal  string
		want    any
		wantErr bool
	}{
		{
			name:    "String value",
			value:   new(string),
			strVal:  "test-value",
			want:    "test-value",
			wantErr: false,
		},
		{
			name:    "Int value",
			value:   new(int),
			strVal:  "42",
			want:    42,
			wantErr: false,
		},
		{
			name:    "Int8 value",
			value:   new(int8),
			strVal:  "127",
			want:    int8(127),
			wantErr: false,
		},
		{
			name:    "Int16 value",
			value:   new(int16),
			strVal:  "32767",
			want:    int16(32767),
			wantErr: false,
		},
		{
			name:    "Int32 value",
			value:   new(int32),
			strVal:  "2147483647",
			want:    int32(2147483647),
			wantErr: false,
		},
		{
			name:    "Int64 value",
			value:   new(int64),
			strVal:  "9223372036854775807",
			want:    int64(9223372036854775807),
			wantErr: false,
		},
		{
			name:    "Uint value",
			value:   new(uint),
			strVal:  "123",
			want:    uint(123),
			wantErr: false,
		},
		{
			name:    "Uint8 value",
			value:   new(uint8),
			strVal:  "255",
			want:    uint8(255),
			wantErr: false,
		},
		{
			name:    "Uint16 value",
			value:   new(uint16),
			strVal:  "65535",
			want:    uint16(65535),
			wantErr: false,
		},
		{
			name:    "Uint32 value",
			value:   new(uint32),
			strVal:  "4294967295",
			want:    uint32(4294967295),
			wantErr: false,
		},
		{
			name:    "Uint64 value",
			value:   new(uint64),
			strVal:  "18446744073709551615",
			want:    uint64(18446744073709551615),
			wantErr: false,
		},
		{
			name:    "Float32 value",
			value:   new(float32),
			strVal:  "3.14",
			want:    float32(3.14),
			wantErr: false,
		},
		{
			name:    "Float64 value",
			value:   new(float64),
			strVal:  "3.141592653589793",
			want:    3.141592653589793,
			wantErr: false,
		},
		{
			name:    "Bool value true",
			value:   new(bool),
			strVal:  "true",
			want:    true,
			wantErr: false,
		},
		{
			name:    "Bool value false",
			value:   new(bool),
			strVal:  "false",
			want:    false,
			wantErr: false,
		},
		{
			name:    "Struct value",
			value:   &struct{ Name string }{},
			strVal:  `{"Name":"test"}`,
			want:    struct{ Name string }{Name: "test"},
			wantErr: false,
		},
		{
			name:    "Slice value",
			value:   &[]string{},
			strVal:  `["a","b","c"]`,
			want:    []string{"a", "b", "c"},
			wantErr: false,
		},
		{
			name:    "Map value",
			value:   &map[string]int{},
			strVal:  `{"a":1,"b":2}`,
			want:    map[string]int{"a": 1, "b": 2},
			wantErr: false,
		},
		{
			name:    "Invalid int format",
			value:   new(int),
			strVal:  "not-a-number",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Invalid int8 format",
			value:   new(int8),
			strVal:  "not-a-number",
			want:    int8(0),
			wantErr: true,
		},
		{
			name:    "Invalid int16 format",
			value:   new(int16),
			strVal:  "not-a-number",
			want:    int16(0),
			wantErr: true,
		},
		{
			name:    "Invalid int32 format",
			value:   new(int32),
			strVal:  "not-a-number",
			want:    int32(0),
			wantErr: true,
		},
		{
			name:    "Invalid int64 format",
			value:   new(int64),
			strVal:  "not-a-number",
			want:    int64(0),
			wantErr: true,
		},
		{
			name:    "Invalid uint format",
			value:   new(uint),
			strVal:  "not-a-number",
			want:    uint(0),
			wantErr: true,
		},
		{
			name:    "Invalid uint8 format",
			value:   new(uint8),
			strVal:  "not-a-number",
			want:    uint8(0),
			wantErr: true,
		},
		{
			name:    "Invalid uint16 format",
			value:   new(uint16),
			strVal:  "not-a-number",
			want:    uint16(0),
			wantErr: true,
		},
		{
			name:    "Invalid uint32 format",
			value:   new(uint32),
			strVal:  "not-a-number",
			want:    uint32(0),
			wantErr: true,
		},
		{
			name:    "Invalid uint64 format",
			value:   new(uint64),
			strVal:  "not-a-number",
			want:    uint64(0),
			wantErr: true,
		},
		{
			name:    "Invalid float format",
			value:   new(float64),
			strVal:  "not-a-number",
			want:    0.0,
			wantErr: true,
		},
		{
			name:    "Invalid float32 format",
			value:   new(float32),
			strVal:  "not-a-number",
			want:    float32(0.0),
			wantErr: true,
		},
		{
			name:    "Invalid bool format",
			value:   new(bool),
			strVal:  "not-a-bool",
			want:    false,
			wantErr: true,
		},
		{
			name:    "Unsupported type",
			value:   new(complex128),
			strVal:  "1+2i",
			want:    complex128(0),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rv := reflect.ValueOf(tt.value)
			err := SetValueByReflect(rv, tt.strVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("setValueByReflectValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got := reflect.ValueOf(tt.value).Elem().Interface()
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("setValueByReflectValue() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestEnvConfigure_GetInvalidValues(t *testing.T) {
	configure := &EnvConfigure{}

	tests := []struct {
		name       string
		key        string
		value      any
		envValue   string
		defaultVal string
		wantErr    bool
	}{
		{
			name:       "Invalid int format",
			key:        "TEST_INVALID_INT",
			value:      new(int),
			envValue:   "not-a-number",
			defaultVal: "0",
			wantErr:    true,
		},
		{
			name:       "Invalid float format",
			key:        "TEST_INVALID_FLOAT",
			value:      new(float64),
			envValue:   "not-a-float",
			defaultVal: "0.0",
			wantErr:    true,
		},
		{
			name:       "Invalid bool format",
			key:        "TEST_INVALID_BOOL",
			value:      new(bool),
			envValue:   "not-a-bool",
			defaultVal: "false",
			wantErr:    true,
		},
		{
			name:       "Invalid duration format",
			key:        "TEST_INVALID_DURATION",
			value:      new(time.Duration),
			envValue:   "not-a-duration",
			defaultVal: "1s",
			wantErr:    true,
		},
		{
			name:       "Invalid JSON for struct",
			key:        "TEST_INVALID_STRUCT",
			value:      &struct{ Name string }{},
			envValue:   "{invalid-json}",
			defaultVal: "{}",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv(GONE+"_"+tt.key, tt.envValue)
			defer os.Unsetenv(GONE + "_" + tt.key)

			err := configure.Get(tt.key, tt.value, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnvConfigure.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvConfigure_GetWithEmptyKey(t *testing.T) {
	configure := &EnvConfigure{}
	value := new(string)

	// Test with empty key but valid default value
	err := configure.Get("", value, "default")
	if err != nil {
		t.Errorf("EnvConfigure.Get() with empty key should use default value, got error: %v", err)
	}
	if *value != "default" {
		t.Errorf("EnvConfigure.Get() with empty key = %v, want %v", *value, "default")
	}
}

func TestEnvConfigure_GetAllNumericTypes(t *testing.T) {
	configure := &EnvConfigure{}
	tests := []struct {
		name       string
		key        string
		value      any
		envValue   string
		defaultVal string
		want       any
		wantErr    bool
	}{
		{
			name:       "int8 value",
			key:        "TEST_INT8",
			value:      new(int8),
			envValue:   "127",
			defaultVal: "0",
			want:       int8(127),
			wantErr:    false,
		},
		{
			name:       "int16 value",
			key:        "TEST_INT16",
			envValue:   "32767",
			value:      new(int16),
			defaultVal: "0",
			want:       int16(32767),
			wantErr:    false,
		},
		{
			name:       "int32 value",
			key:        "TEST_INT32",
			envValue:   "2147483647",
			value:      new(int32),
			defaultVal: "0",
			want:       int32(2147483647),
			wantErr:    false,
		},
		{
			name:       "uint8 value",
			key:        "TEST_UINT8",
			envValue:   "255",
			value:      new(uint8),
			defaultVal: "0",
			want:       uint8(255),
			wantErr:    false,
		},
		{
			name:       "uint16 value",
			key:        "TEST_UINT16",
			envValue:   "65535",
			value:      new(uint16),
			defaultVal: "0",
			want:       uint16(65535),
			wantErr:    false,
		},
		{
			name:       "uint32 value",
			key:        "TEST_UINT32",
			envValue:   "4294967295",
			value:      new(uint32),
			defaultVal: "0",
			want:       uint32(4294967295),
			wantErr:    false,
		},
		{
			name:       "float32 value",
			key:        "TEST_FLOAT32",
			envValue:   "3.14159",
			value:      new(float32),
			defaultVal: "0",
			want:       float32(3.14159),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(GONE+"_"+tt.key, tt.envValue)
			defer os.Unsetenv(GONE + "_" + tt.key)

			err := configure.Get(tt.key, tt.value, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnvConfigure.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got := reflect.ValueOf(tt.value).Elem().Interface()
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("EnvConfigure.Get() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestEnvConfigure_GetEmptyEnvWithDefaults(t *testing.T) {
	configure := &EnvConfigure{}
	tests := []struct {
		name       string
		key        string
		value      any
		defaultVal string
		want       any
		wantErr    bool
	}{
		{
			name:       "Empty int with default",
			key:        "NONEXISTENT_INT",
			value:      new(int),
			defaultVal: "42",
			want:       42,
			wantErr:    false,
		},
		{
			name:       "Empty float64 with default",
			key:        "NONEXISTENT_FLOAT",
			value:      new(float64),
			defaultVal: "3.14",
			want:       3.14,
			wantErr:    false,
		},
		{
			name:       "Empty bool with default",
			key:        "NONEXISTENT_BOOL",
			value:      new(bool),
			defaultVal: "true",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "Empty duration with default",
			key:        "NONEXISTENT_DURATION",
			value:      new(time.Duration),
			defaultVal: "1h",
			want:       time.Hour,
			wantErr:    false,
		},
		{
			name:       "Empty string with empty default",
			key:        "NONEXISTENT_STRING",
			value:      new(string),
			defaultVal: "",
			want:       "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := configure.Get(tt.key, tt.value, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnvConfigure.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got := reflect.ValueOf(tt.value).Elem().Interface()
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("EnvConfigure.Get() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestConfigProvider_ProvidePointerType(t *testing.T) {
	mockConfigure := &SimpleMockConfigure{
		values: map[string]string{
			"ptr-string": "test-value",
		},
	}

	provider := &ConfigProvider{
		configure: mockConfigure,
	}

	tests := []struct {
		name      string
		tagConf   string
		valueType reflect.Type
		want      any
		wantErr   bool
	}{
		{
			name:      "Pointer string type",
			tagConf:   "ptr-string",
			valueType: reflect.PtrTo(reflect.TypeOf("")),
			want:      func() any { s := "test-value"; return &s }(),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.Provide(tt.tagConf, tt.valueType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigProvider.Provide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ConfigProvider.Provide() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestEnvConfigure_GetInvalidNumericValues(t *testing.T) {
	configure := &EnvConfigure{}
	tests := []struct {
		name       string
		key        string
		value      any
		envValue   string
		defaultVal string
		wantErr    bool
	}{
		{
			name:       "Invalid int8 overflow",
			key:        "TEST_INT8",
			value:      new(int8),
			envValue:   "128",
			defaultVal: "0",
			wantErr:    true,
		},
		{
			name:       "Invalid int16 overflow",
			key:        "TEST_INT16",
			value:      new(int16),
			envValue:   "32768",
			defaultVal: "0",
			wantErr:    true,
		},
		{
			name:       "Invalid uint8 overflow",
			key:        "TEST_UINT8",
			value:      new(uint8),
			envValue:   "256",
			defaultVal: "0",
			wantErr:    true,
		},
		{
			name:       "Invalid uint16 overflow",
			key:        "TEST_UINT16",
			value:      new(uint16),
			envValue:   "65536",
			defaultVal: "0",
			wantErr:    true,
		},
		{
			name:       "Invalid negative uint",
			key:        "TEST_UINT",
			value:      new(uint),
			envValue:   "-1",
			defaultVal: "0",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(GONE+"_"+tt.key, tt.envValue)
			defer os.Unsetenv(GONE + "_" + tt.key)

			err := configure.Get(tt.key, tt.value, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnvConfigure.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigProvider_ProvideMultipleKeys(t *testing.T) {
	mockConfigure := &SimpleMockConfigure{
		values: map[string]string{
			"second-key": "found-value",
		},
	}

	provider := &ConfigProvider{
		configure: mockConfigure,
	}

	tests := []struct {
		name      string
		tagConf   string
		valueType reflect.Type
		want      any
		wantErr   bool
	}{
		{
			name:      "Multiple keys with second key found",
			tagConf:   "first-key,second-key",
			valueType: reflect.TypeOf(""),
			want:      "",
			wantErr:   false,
		},
		{
			name:      "Multiple keys with default",
			tagConf:   "missing-key1,missing-key2,default=default-value",
			valueType: reflect.TypeOf(""),
			want:      "default-value",
			wantErr:   false,
		},
		{
			name:      "Multiple keys all missing no default",
			tagConf:   "missing-key1,missing-key2",
			valueType: reflect.TypeOf(""),
			want:      "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.Provide(tt.tagConf, tt.valueType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigProvider.Provide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConfigProvider.Provide() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigProvider_ProvideInvalidJSON(t *testing.T) {
	mockConfigure := &SimpleMockConfigure{
		values: map[string]string{
			"invalid-json": "{invalid-json}",
		},
	}

	provider := &ConfigProvider{
		configure: mockConfigure,
	}

	type testStruct struct {
		Name string `json:"name"`
	}

	_, err := provider.Provide("invalid-json", reflect.TypeOf(testStruct{}))
	if err == nil {
		t.Error("ConfigProvider.Provide() should return error for invalid JSON")
	}
}

func TestEnvConfigure_GetEmptyValue(t *testing.T) {
	configure := &EnvConfigure{}
	tests := []struct {
		name       string
		key        string
		value      any
		defaultVal string
		want       any
		wantErr    bool
	}{
		{
			name:       "Empty value for int",
			key:        "EMPTY_INT",
			value:      new(int),
			defaultVal: "",
			want:       0,
			wantErr:    false,
		},
		{
			name:       "Empty value for bool",
			key:        "EMPTY_BOOL",
			value:      new(bool),
			defaultVal: "",
			want:       false,
			wantErr:    false,
		},
		{
			name:       "Empty value for duration",
			key:        "EMPTY_DURATION",
			value:      new(time.Duration),
			defaultVal: "",
			want:       time.Duration(0),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(GONE+"_"+tt.key, "")
			defer os.Unsetenv(GONE + "_" + tt.key)

			err := configure.Get(tt.key, tt.value, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnvConfigure.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got := reflect.ValueOf(tt.value).Elem().Interface()
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("EnvConfigure.Get() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestEnvConfigure_GetWithNilValue(t *testing.T) {
	configure := &EnvConfigure{}
	err := configure.Get("TEST_KEY", nil, "default")
	if err == nil {
		t.Error("EnvConfigure.Get() should return error for nil value")
	}
}

func TestConfigProvider_ProvideWithInvalidType(t *testing.T) {
	mockConfigure := &SimpleMockConfigure{
		values: map[string]string{
			"test-key": "test-value",
		},
	}

	provider := &ConfigProvider{
		configure: mockConfigure,
	}

	// Test with invalid type (chan)
	_, err := provider.Provide("test-key", reflect.TypeOf(make(chan int)))
	if err == nil {
		t.Error("ConfigProvider.Provide() should return error for invalid type")
	}
}

func TestEnvConfigure_GetWithComplexTypes(t *testing.T) {
	configure := &EnvConfigure{}
	tests := []struct {
		name       string
		key        string
		value      any
		envValue   string
		defaultVal string
		want       any
		wantErr    bool
	}{
		{
			name:       "Complex struct with nested fields",
			key:        "TEST_COMPLEX",
			value:      &struct{ Inner struct{ Value int } }{},
			envValue:   `{"Inner":{"Value":42}}`,
			defaultVal: "{}",
			want:       struct{ Inner struct{ Value int } }{Inner: struct{ Value int }{Value: 42}},
			wantErr:    false,
		},
		{
			name:       "Array type",
			key:        "TEST_ARRAY",
			value:      &[]int{},
			envValue:   "[1,2,3]",
			defaultVal: "[]",
			want:       []int{1, 2, 3},
			wantErr:    false,
		},
		{
			name:       "Map type",
			key:        "TEST_MAP",
			value:      &map[string]int{},
			envValue:   `{"key":42}`,
			defaultVal: "{}",
			want:       map[string]int{"key": 42},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(GONE+"_"+tt.key, tt.envValue)
			defer os.Unsetenv(GONE + "_" + tt.key)

			err := configure.Get(tt.key, tt.value, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnvConfigure.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got := reflect.ValueOf(tt.value).Elem().Interface()
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("EnvConfigure.Get() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestConfigProvider_ProvideWithEmptyKeys(t *testing.T) {
	mockConfigure := &SimpleMockConfigure{
		values: map[string]string{},
	}

	provider := &ConfigProvider{
		configure: mockConfigure,
	}

	tests := []struct {
		name      string
		tagConf   string
		valueType reflect.Type
		wantErr   bool
	}{
		{
			name:      "Empty tag config",
			tagConf:   "",
			valueType: reflect.TypeOf(""),
			wantErr:   true,
		},
		{
			name:      "Only default value",
			tagConf:   "default=test",
			valueType: reflect.TypeOf(""),
			wantErr:   false,
		},
		{
			name:      "Only comma",
			tagConf:   ",",
			valueType: reflect.TypeOf(""),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := provider.Provide(tt.tagConf, tt.valueType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigProvider.Provide() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvConfigure_GetWithInvalidDefaultValue(t *testing.T) {
	configure := &EnvConfigure{}
	tests := []struct {
		name       string
		key        string
		value      any
		defaultVal string
		wantErr    bool
	}{
		{
			name:       "Invalid default int",
			key:        "TEST_INT_X",
			value:      new(int),
			defaultVal: "not-a-number",
			wantErr:    true,
		},
		{
			name:       "Invalid default duration",
			key:        "TEST_DURATION_X",
			value:      new(time.Duration),
			defaultVal: "invalid-duration",
			wantErr:    true,
		},
		{
			name:       "Invalid default JSON",
			key:        "TEST_STRUCT_X",
			value:      &struct{ Value int }{},
			defaultVal: "{invalid-json}",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := configure.Get(tt.key, tt.value, tt.defaultVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnvConfigure.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
