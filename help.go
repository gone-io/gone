package gone

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// PanicTrace 用于获取调用者的堆栈信息
func PanicTrace(kb int, skip int) []byte {
	stack := make([]byte, kb<<10) //4KB
	length := runtime.Stack(stack, true)

	_, filename, fileLine, ok := runtime.Caller(skip)
	start := 0
	if ok {
		start = bytes.Index(stack, []byte(fmt.Sprintf("%s:%d", filename, fileLine)))
		stack = stack[start:length]
	}

	line := []byte("\n")
	start = bytes.Index(stack, line) + 1
	stack = stack[start:]
	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}

	e := []byte("\ngoroutine ")
	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")
	return stack
}

// GetFuncName 获取某个函数的名字
func GetFuncName(f any) string {
	return strings.TrimSuffix(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "-fm")
}

// GetInterfaceType 获取接口的类型
func GetInterfaceType[T any](t *T) reflect.Type {
	return reflect.TypeOf(t).Elem()
}

func WrapNormalFnToProcess(fn any) Process {
	return func(cemetery Cemetery) error {
		args, err := cemetery.InjectFuncParameters(fn, nil, nil)
		if err != nil {
			return err
		}

		var parameters []reflect.Value
		for _, arg := range args {
			parameters = append(parameters, reflect.ValueOf(arg))
		}

		results := reflect.ValueOf(fn).Call(parameters)
		for _, result := range results {
			if err, ok := result.Interface().(error); ok {
				return err
			}
		}
		return nil
	}
}

// IsCompatible t Type can put in goner
func IsCompatible(t reflect.Type, goner Goner) bool {
	gonerType := reflect.TypeOf(goner)

	switch t.Kind() {
	case reflect.Interface:
		return gonerType.Implements(t)
	case reflect.Struct:
		return gonerType.Elem() == t
	default:
		return gonerType == t
	}
}

func (c *cemetery) setFieldValue(v reflect.Value, ref any) {
	t := v.Type()

	switch t.Kind() {
	case reflect.Interface, reflect.Pointer, reflect.Slice, reflect.Map:
		v.Set(reflect.ValueOf(ref))
	default:
		v.Set(reflect.ValueOf(ref).Elem())
	}
	return
}

type timeUseRecord struct {
	Count   int64
	UseTime time.Duration
}

var mapRecord = make(map[string]*timeUseRecord)

func TimeStat(name string, start time.Time, logs ...func(format string, args ...any)) {
	since := time.Since(start)
	if mapRecord[name] == nil {
		mapRecord[name] = &timeUseRecord{}
	}
	mapRecord[name].UseTime += since
	mapRecord[name].Count++

	var log func(format string, args ...any)
	if len(logs) == 0 {
		log = func(format string, args ...any) {
			fmt.Printf(format, args...)
		}
	} else {
		log = logs[0]
	}

	log("%s count %v, use %v, avg: %v\n",
		name,
		mapRecord[name].Count,
		mapRecord[name].UseTime,
		mapRecord[name].UseTime/time.Duration(mapRecord[name].Count),
	)
}
