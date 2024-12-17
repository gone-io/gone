package gone

import (
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
		if t.Name() != "" {
			if t.PkgPath() != "" {
				return t.PkgPath() + "." + t.Name()
			}
			return t.Name()
		}
		return t.String()
	}
}

// GetFuncName get function name
func GetFuncName(f any) string {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return ""
	}

	if t.Name() != "" {
		if t.PkgPath() != "" {
			return t.PkgPath() + "." + t.Name()
		}
		return t.Name()
	}

	// Fallback to runtime.FuncForPC for anonymous functions
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

func OnceLoad(fn LoadFunc) LoadFunc {
	var key = GenLoaderKey()
	return func(loader Loader) error {
		if loader.Loaded(key) {
			return nil
		}
		return fn(loader)
	}
}

// SafeExecute 执行可能会触发panic的函数并将panic转换为error
func SafeExecute(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = NewInnerErrorSkip(fmt.Sprintf("panic occurred: %v", r), FailInstall, 7)
		}
	}()
	// 执行传入的函数
	fn()
	return nil
}

func convertUppercaseCamel(input string) string {
	parts := strings.Split(input, ".")
	for i, part := range parts {
		parts[i] = strings.ToUpper(part)
	}
	return strings.Join(parts, "_")
}
