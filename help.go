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
func InjectWrapFn(cemetery Cemetery, fn any) (*reflect.Value, error) {
	return InjectWrapFnWithHook(cemetery, fn, nil, nil)
}

func InjectWrapFnWithHook(cemetery Cemetery, fn any, before func([]reflect.Value), after func([]reflect.Value)) (*reflect.Value, error) {
	ft := reflect.TypeOf(fn)
	fv := reflect.ValueOf(fn)
	if ft.Kind() != reflect.Func {
		return nil, NewInnerError("fn must be a function", NotCompatible)
	}

	in := ft.NumIn()
	if in > 1 {
		return nil, NewInnerError("fn only support one input parameter or no input parameter", NotCompatible)
	}

	var args = make([]reflect.Value, 0)

	if in == 1 {
		if ft.In(0).Kind() != reflect.Struct {
			return nil, NewInnerError("fn input parameter must be a struct", NotCompatible)
		}

		pt := ft.In(0)

		if pt.Name() != "" || pt.PkgPath() != "" {
			return nil, NewInnerError("fn input parameter must be a anonymous struct", NotCompatible)
		}

		parameter := reflect.New(pt)

		goner := parameter.Interface()
		_, err := cemetery.ReviveOne(goner)
		if err != nil {
			return nil, ToError(err)
		}
		args = append(args, parameter.Elem())
	}

	var outList []reflect.Type
	for i := 0; i < ft.NumOut(); i++ {
		outList = append(outList, ft.Out(i))
	}

	makeFunc := reflect.MakeFunc(reflect.FuncOf(nil, outList, false), func([]reflect.Value) (results []reflect.Value) {
		if before != nil {
			before(args)
		}
		results = fv.Call(args)
		if after != nil {
			after(results)
		}
		return
	})
	return &makeFunc, nil
}

func ExecuteInjectWrapFn(fn *reflect.Value) (results []any) {
	call := fn.Call([]reflect.Value{})

	for i := 0; i < len(call); i++ {
		arg := call[i].Interface()
		v := reflect.ValueOf(arg)

		if v.Kind() == reflect.Pointer && v.IsNil() {
			results = append(results, nil)
		} else {
			results = append(results, arg)
		}
	}
	return
}

func WrapNormalFnToProcess(fn any) Process {
	return func(cemetery Cemetery) error {
		wrapFn, err := InjectWrapFn(cemetery, fn)
		if err != nil {
			return err
		}
		results := ExecuteInjectWrapFn(wrapFn)
		for _, result := range results {
			if err, ok := result.(error); ok {
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

func TimeStat(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s use %v\n", name, time.Since(start))
	}
}
