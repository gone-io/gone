package gone

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
)

// MockConfigure implements Configure interface for testing
type MockConfigure struct {
	values map[string]string
}

func (m *MockConfigure) Get(key string, v any, defaultVal string) error {
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
	mockConfigure := &MockConfigure{
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
		"TEST_STRING":   "test-value",
		"TEST_INT":      "42",
		"TEST_INT64":    "9223372036854775807",
		"TEST_FLOAT":    "3.14",
		"TEST_BOOL":     "true",
		"TEST_UINT":     "123",
		"TEST_UINT64":   "18446744073709551615",
		"TEST_DURATION": "1h30m",
		"TEST_STRUCT":   `{"name":"test","value":123}`,
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
	mockConfigure := &MockConfigure{
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
