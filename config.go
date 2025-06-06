package gone

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
)

const ConfigureName = "configure"

// Configure defines the interface for configuration providers
// Get retrieves a configuration value by key, storing it in v, with a default value if not found
type Configure interface {
	Get(key string, v any, defaultVal string) error
}

type DynamicConfigure interface {
	Configure
	Notify(key string, callback ConfWatchFunc)
}

type ConfWatchFunc func(oldVal, newVal any)

type ConfWatcher func(key string, callback ConfWatchFunc)

type confWatcherProvider struct {
	Flag
	logger    Logger    `gone:"*"`
	configure Configure `gone:"configure"`
	m         map[string][]ConfWatchFunc
}

func (p *confWatcherProvider) Init() {
	p.m = make(map[string][]ConfWatchFunc)
}

func (p *confWatcherProvider) Provide(string) (ConfWatcher, error) {
	if configure, ok := p.configure.(DynamicConfigure); !ok {
		p.logger.Warnf("configure(%T) is not DynamicConfigure, cannot support watch", configure)
		return nil, nil
	} else {
		return func(key string, callback ConfWatchFunc) {
			p.m[key] = append(p.m[key], callback)
			configure.Notify(key, func(oldVal, newVal any) {
				funcs := p.m[key]
				for _, f := range funcs {
					err := SafeExecute(func() error {
						f(oldVal, newVal)
						return nil
					})
					if err != nil {
						p.logger.Warnf("call %T err:%v", f, err)
					}
				}
			})
		}, nil
	}
}

// ConfigProvider implements a provider for injecting configuration values
// It uses an underlying Configure implementation to retrieve values
type ConfigProvider struct {
	Flag
	configure Configure `gone:"configure"` // The Configure implementation to use
	mu        sync.RWMutex
}

// GonerName returns the provider name "config" used for registration
func (s *ConfigProvider) GonerName() string {
	return "config"
}

func (s *ConfigProvider) Init() {}

// Provide implements the provider interface to inject configuration values
// Parameters:
//   - tagConf: The tag configuration string containing key and default value
//   - t: The reflect.Type of the value to provide
//
// Returns:
//   - The configured value of type t
//   - Error if configuration fails
func (s *ConfigProvider) Provide(tagConf string, t reflect.Type) (any, error) {
	// Parse the tag string into a map and ordered keys
	m, keys := TagStringParse(tagConf)
	if len(keys) == 0 || len(keys) == 1 && keys[0] == "" {
		return nil, NewInnerError("config-key is empty", ConfigError)
	}

	// Get the first key and its default value
	key := keys[0]
	defaultValue := m[key]
	if defaultValue == "" {
		defaultValue = m["default"] // Fallback to "default" key if no value
	}

	var getType = t
	if t.Kind() == reflect.Ptr {
		getType = t.Elem()
	}

	// Create new value of requested type and configure it
	value := reflect.New(getType)
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.configure.Get(key, value.Interface(), defaultValue); err != nil {
		return nil, ToError(err)
	}
	if t.Kind() == reflect.Ptr {
		return value.Interface(), nil
	}
	return value.Elem().Interface(), nil
}

type EnvConfigure struct {
	Flag
}

const GONE = "GONE"

// Get retrieves a configuration value from environment variables with fallback to default value.
// Supports type conversion for various Go types including string, int, float, bool, and structs.
//
// Parameters:
//   - key: Environment variable name to look up
//   - v: Pointer to variable where the value will be stored
//   - defaultVal: Default value if environment variable is not set
//
// Returns error if:
//   - v is not a pointer
//   - Type conversion fails
//   - Unsupported type is provided
func (s *EnvConfigure) Get(key string, v any, defaultVal string) error {
	// Get environment variable value, fallback to default if not set
	key = convertUppercaseCamel(GONE + "_" + key)
	env := os.Getenv(key)
	if env == "" {
		env = defaultVal
	}

	// Verify v is a pointer
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return NewInnerError("Value must be a pointer", ConfigError)
	}

	// Set default "0" for numeric and boolean types when env is empty
	if env == "" {
		switch v.(type) {
		case *int, *int8, *int16, *int32, *int64,
			*uint, *uint8, *uint16, *uint32, *uint64,
			*float32, *float64, *bool, *time.Duration:
			env = "0"
		}
	}
	return SetValue(rv, v, env)
}

var UnsupportedError = NewInnerError("Unsupported type by EnvConfigure", ConfigError)

// SetValue sets the value of a pointer to a Go type based on the provided value and environment variable.
// Deprecated use SetPointerValue or SetValueByReflect instead
func SetValue(rv reflect.Value, v any, value string) error {
	err := SetPointerValue(v, value)
	if err != nil {
		if errors.Is(err, UnsupportedError) {
			return SetValueByReflect(rv, value)
		} else {
			return ToError(err)
		}
	}
	return nil
}

// SetPointerValue sets the value of a pointer to a Go type based on the provided value and environment variable.
func SetPointerValue(v any, value string) error {
	// Type switch to handle different pointer types
	switch ptr := v.(type) {
	// String type
	case *string:
		*ptr = value

	// Int types
	case *int:
		val, err := strconv.Atoi(value)
		if err != nil {
			return ToError(err)
		}
		*ptr = val
	case *int8:
		val, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return ToError(err)
		}
		*ptr = int8(val)
	case *int16:
		val, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return ToError(err)
		}
		*ptr = int16(val)
	case *int32:
		val, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return ToError(err)
		}
		*ptr = int32(val)
	case *int64:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ToError(err)
		}
		*ptr = val

	// Unsigned int types
	case *uint:
		val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return ToError(err)
		}
		*ptr = uint(val)
	case *uint8:
		val, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return ToError(err)
		}
		*ptr = uint8(val)
	case *uint16:
		val, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return ToError(err)
		}
		*ptr = uint16(val)
	case *uint32:
		val, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return ToError(err)
		}
		*ptr = uint32(val)
	case *uint64:
		val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return ToError(err)
		}
		*ptr = val

	// Float types
	case *float32:
		val, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return ToError(err)
		}
		*ptr = float32(val)
	case *float64:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return ToError(err)
		}
		*ptr = val

	// Boolean type
	case *bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return ToError(err)
		}
		*ptr = val

	// Time duration type
	case *time.Duration:
		val, err := time.ParseDuration(value)
		if err != nil {
			return ToError(err)
		}
		*ptr = val

	default:
		return UnsupportedError
	}
	return nil
}

// SetValueByReflect sets the value of a pointer to a Go type based on the provided value and environment variable.
func SetValueByReflect(rv reflect.Value, value string) error {
	k := rv.Elem().Kind()
	switch k {
	case reflect.Struct, reflect.Slice, reflect.Map:
		if value != "" {
			return ToError(json.Unmarshal([]byte(value), rv.Interface()))
		}
	case reflect.String:
		rv.Elem().SetString(value)
	case reflect.Int:
		val, err := strconv.Atoi(value)
		if err != nil {
			return ToError(err)
		}
		rv.Elem().SetInt(int64(val))
	case reflect.Int8:
		val, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return ToError(err)
		}
		rv.Elem().SetInt(val)
	case reflect.Int16:
		val, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return ToError(err)
		}
		rv.Elem().SetInt(val)
	case reflect.Int32:
		val, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return ToError(err)
		}
		rv.Elem().SetInt(val)
	case reflect.Int64:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ToError(err)
		}
		rv.Elem().SetInt(val)

	case reflect.Uint:
		val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return ToError(err)
		}
		rv.Elem().SetUint(val)
	case reflect.Uint8:
		val, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return ToError(err)
		}
		rv.Elem().SetUint(val)
	case reflect.Uint16:
		val, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return ToError(err)
		}
		rv.Elem().SetUint(val)
	case reflect.Uint32:
		val, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return ToError(err)
		}
		rv.Elem().SetUint(val)
	case reflect.Uint64:
		val, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return ToError(err)
		}
		rv.Elem().SetUint(val)

	case reflect.Float32:
		val, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return ToError(err)
		}
		rv.Elem().SetFloat(val)
	case reflect.Float64:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return ToError(err)
		}
		rv.Elem().SetFloat(val)

	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return ToError(err)
		}
		rv.Elem().SetBool(val)

	default:
		return UnsupportedError
	}
	return nil
}
