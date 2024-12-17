package gone

import (
	"github.com/gone-io/gone/internal/json"
	"os"
	"reflect"
	"strconv"
	"time"
)

const ConfigureName = "configure"

// Configure defines the interface for configuration providers
// Get retrieves a configuration value by key, storing it in v, with a default value if not found
type Configure interface {
	Get(key string, v any, defaultVal string) error
}

// ConfigProvider implements a provider for injecting configuration values
// It uses an underlying Configure implementation to retrieve values
type ConfigProvider struct {
	Flag
	configure Configure `gone:"configure"` // The Configure implementation to use
}

// Name returns the provider name "config" used for registration
func (s *ConfigProvider) Name() string {
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
	if len(keys) == 0 {
		return nil, NewInnerError("config-key is empty", ConfigError)
	}

	// Get the first key and its default value
	key := keys[0]
	defaultValue := m[key]
	if defaultValue == "" {
		defaultValue = m["default"] // Fallback to "default" key if no value
	}

	// Create new value of requested type and configure it
	value := reflect.New(t)
	err := s.configure.Get(key, value.Interface(), defaultValue)
	if err != nil {
		return nil, ToError(err)
	}
	return value.Elem().Interface(), nil
}

type EnvConfigure struct {
	Flag
}

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
	key = convertUppercaseCamel("GONE_" + key)
	env := os.Getenv(key)
	if env == "" {
		env = defaultVal
	}

	// Verify v is a pointer
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return NewInnerError("Value must be a pointer", ConfigError)
	}

	// Type switch to handle different pointer types
	switch ptr := v.(type) {
	case *string:
		// String type needs no conversion
		*ptr = env
	case *int:
		// Convert string to int
		val, err := strconv.Atoi(env)
		if err != nil {
			return ToError(err)
		}
		*ptr = val
	case *int64:
		// Convert string to int64
		val, err := strconv.ParseInt(env, 10, 64)
		if err != nil {
			return ToError(err)
		}
		*ptr = val
	case *float64:
		// Convert string to float64
		val, err := strconv.ParseFloat(env, 64)
		if err != nil {
			return ToError(err)
		}
		*ptr = val
	case *bool:
		// Convert string to bool
		val, err := strconv.ParseBool(env)
		if err != nil {
			return ToError(err)
		}
		*ptr = val
	case *uint:
		// Convert string to uint
		val, err := strconv.ParseUint(env, 10, 64)
		if err != nil {
			return ToError(err)
		}
		*ptr = uint(val)
	case *uint64:
		// Convert string to uint64
		val, err := strconv.ParseUint(env, 10, 64)
		if err != nil {
			return ToError(err)
		}
		*ptr = val
	case *time.Duration:
		// Convert string to time.Duration
		val, err := time.ParseDuration(env)
		if err != nil {
			return ToError(err)
		}
		*ptr = val
	default:
		// Handle struct types by JSON unmarshal
		if rv.Elem().Kind() == reflect.Struct {
			err := json.Unmarshal([]byte(env), v)
			if err != nil {
				return ToError(err)
			}
			return nil
		}
		return NewInnerError("Unsupported type by EnvConfigure", ConfigError)
	}
	return nil
}
