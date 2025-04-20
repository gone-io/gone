package gone

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"unsafe"
)

// TagStringParse parses a tag string in the format "key1=value1,key2=value2" into a map and ordered key slice.
// It splits the string by commas, then splits each part into key-value pairs by "=".
// Returns:
//   - map[string]string: Contains all key-value pairs from the tag string
//   - []string: Contains keys in order of appearance, with duplicates removed
func TagStringParse(conf string) (map[string]string, []string) {
	conf = strings.TrimSpace(conf)
	specs := strings.Split(conf, ",")
	m := make(map[string]string)
	var keys []string
	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		pairs := strings.Split(spec, "=")
		var key, value string
		if len(pairs) > 0 {
			key = strings.TrimSpace(pairs[0])
		}
		if len(pairs) > 1 {
			value = strings.TrimSpace(pairs[1])
		}
		if _, ok := m[key]; !ok {
			keys = append(keys, key)
		}
		m[key] = value
	}
	return m, keys
}

// ParseGoneTag parses a gone tag string in the format "name,extend" into name and extend parts.
// The name part is used to identify the goner, while extend part contains additional configuration.
// For example:
//   - "myGoner" returns ("myGoner", "")
//   - "myGoner,config=value" returns ("myGoner", "config=value")
func ParseGoneTag(tag string) (name string, extend string) {
	if tag == "" {
		return
	}
	list := strings.SplitN(tag, ",", 2)
	switch len(list) {
	case 1:
		name = list[0]
	default:
		name, extend = list[0], list[1]
	}
	return
}

func BlackMagic(v reflect.Value) reflect.Value {
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}

// IsCompatible checks if a goner object is compatible with a given type t.
// For interface types, checks if goner implements the interface.
// For other types, checks for exact type equality.
func IsCompatible(t reflect.Type, goner any) bool {
	gonerType := reflect.TypeOf(goner)

	switch t.Kind() {
	case reflect.Interface:
		return gonerType.Implements(t)
	//case reflect.Struct:
	//	return gonerType.Elem() == t
	default:
		return gonerType == t
	}
}

// GetTypeName returns a string representation of a reflect.Type, including package path for named types.
// For arrays, slices, maps and pointers it recursively formats the element types.
// For interfaces and structs it includes the package path if available.
// For unnamed types it returns a basic representation like "interface{}" or "struct{}".
func GetTypeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), GetTypeName(t.Elem()))
	case reflect.Slice:
		return "[]" + GetTypeName(t.Elem())
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", GetTypeName(t.Key()), GetTypeName(t.Elem()))
	case reflect.Ptr:
		return "*" + GetTypeName(t.Elem())
	case reflect.Interface:
		if t.Name() != "" {
			return t.PkgPath() + "." + t.Name()
		}
		return "interface{}"
	case reflect.Struct:
		if t.Name() != "" {
			return t.PkgPath() + "." + t.Name()
		}
		return "struct{}"
	default:
		name := t.Name()
		if name == "" {
			name = t.String()
		}
		if t.PkgPath() != "" {
			return t.PkgPath() + "." + name
		}
		return name
	}
}

// GetFuncName get function name
func GetFuncName(f any) string {
	of := reflect.ValueOf(f)
	t := of.Type()
	if t.Kind() != reflect.Func {
		return ""
	}
	return strings.TrimSuffix(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "-fm")
}

// RemoveRepeat removes duplicate pointers from a slice of pointers to type T.
// It preserves the order of first occurrence of each pointer.
func RemoveRepeat[T comparable](list []T) []T {
	type X struct{}
	m := make(map[T]X)
	out := make([]T, 0, len(list))
	for _, v := range list {
		if _, ok := m[v]; !ok {
			m[v] = X{}
			out = append(out, v)
		}
	}
	return out
}

// GenLoaderKey will return a brand new, never-before-used LoaderKey
func GenLoaderKey() LoaderKey {
	keyMtx.Lock()
	defer keyMtx.Unlock()
	keyCounter += 1
	return LoaderKey{id: keyCounter}
}

var loadFuncMap = make(map[string]LoaderKey)

func genLoaderKey(fn any) LoaderKey {
	key := fmt.Sprintf("%#v", fn)
	if k, ok := loadFuncMap[key]; ok {
		return k
	} else {
		loadFuncMap[key] = GenLoaderKey()
		return loadFuncMap[key]
	}
}

// OnceLoad wraps a LoadFunc to ensure it only executes once per Loader instance.
// It generates a unique LoaderKey for the function and uses it to track execution status.
//
// Parameters:
//   - fn: The LoadFunc to be wrapped. This function will only be executed once per Loader.
//
// Returns:
//   - LoadFunc: A wrapped function that checks if it has already been executed before calling the original function.
//
// Example usage:
// ```go
//
//	func loadComponents(l Loader) error {
//	    // Load dependencies...
//	    return nil
//	}
//
//	// Create a function that will only execute once per Loader
//	wrappedLoad := OnceLoad(loadComponents)
//
//	// First call executes loadComponents
//	wrappedLoad(loader)
//
//	// Second call returns nil without executing loadComponents again
//	wrappedLoad(loader)
//
// ```
func OnceLoad(fn LoadFunc) LoadFunc {
	return func(loader Loader) error {
		var key = genLoaderKey(fn)
		if loader.Loaded(key) {
			return nil
		}
		return fn(loader)
	}
}

// BuildSingProviderLoadFunc creates a LoadFunc that wraps a FunctionProvider and ensures it's loaded only once per Loader instance.
// It combines the functionality of OnceLoad and WrapFunctionProvider to create a reusable loader function.
//
// Parameters:
//   - fn: The FunctionProvider to be wrapped. This function will be converted to a provider component.
//   - options: Optional configuration for how the provider should be loaded.
//
// Returns:
//   - LoadFunc: A wrapped function that loads the provider only once per Loader instance.
//
// Example usage:
// ```go
//
//	func createService(config string, param struct {
//			receiveInjected InjectedRepo `gone:"*"`
//		}) (Service, error) {
//		// Create and return a service instance using the config and repository
//		return &ServiceImpl{repo: param.receiveInjected, config: config}, nil
//	}
//
// // Create a loader function that will only load the service provider once
// serviceLoader := BuildSingProviderLoadFunc(createService)
// // Load the service provider into the container
// NewApp().Loads(serviceLoader)
//
// ```
func BuildSingProviderLoadFunc[P, T any](fn FunctionProvider[P, T], options ...Option) LoadFunc {
	return OnceLoad(func(loader Loader) error {
		provider := WrapFunctionProvider(fn)
		return loader.Load(provider, options...)
	})
}

// BuildThirdComponentLoadFunc creates a LoadFunc that registers an existing component into the container.
// It wraps the component in a simple provider function and ensures it's loaded only once per Loader instance.
//
// Parameters:
//   - component: The existing component instance to be registered in the container.
//   - options: Optional configuration for how the component should be loaded.
//
// Returns:
//   - LoadFunc: A wrapped function that loads the component only once per Loader instance.
//
// Example usage:
// ```go
//
//	type TestComponent struct {
//		i int
//	}
//	// Create an existing component instance
//	var component TestComponent
//
//	// Create a loader function that will register the component in the container
//	loadFunc := BuildThirdComponentLoadFunc(&component)
//
//	// Load the component into the container
//	NewApp().Loads(loadFunc)
//
// ```
func BuildThirdComponentLoadFunc[T any](component T, options ...Option) LoadFunc {
	return BuildSingProviderLoadFunc(func(tagConf string, param struct{}) (T, error) {
		return component, nil
	}, options...)
}

// SafeExecute safely executes a function and captures any panics that occur during execution.
// It converts panics into error returns, allowing for graceful error handling in code that might panic.
//
// Parameters:
//   - fn: The function to execute safely. This function should return an error or nil.
//
// Returns:
//   - error: The error returned by fn, or a new InnerError if a panic occurred during execution.
//
// Example usage:
// ```go
//
//	func riskyOperation() error {
//	    // Code that might panic
//	    return nil
//	}
//
//	// Execute the risky operation safely
//	err := SafeExecute(riskyOperation)
//	if err != nil {
//	    // Handle the error or panic gracefully
//	}
//
// ```
func SafeExecute(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = NewInnerErrorSkip(fmt.Sprintf("panic occurred: %v", r), FailInstall, 3)
		}
	}()
	return fn()
}

func convertUppercaseCamel(input string) string {
	parts := strings.Split(input, ".")
	for i, part := range parts {
		parts[i] = strings.ToUpper(part)
	}
	return strings.Join(parts, "_")
}

// GetInterfaceType get interface type
func GetInterfaceType[T any](t *T) reflect.Type {
	return reflect.TypeOf(t).Elem()
}

func IsError(err error, code int) bool {
	var e Error
	if errors.As(err, &e) {
		return e.Code() == code
	}
	return false
}
