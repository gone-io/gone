package gone

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// PanicTrace 用于获取调用者的堆栈信息
func PanicTrace(kb int) []byte {
	e := []byte("\ngoroutine ")
	line := []byte("\n")
	stack := make([]byte, kb<<10) //4KB
	length := runtime.Stack(stack, true)

	_, filename, fileLine, ok := runtime.Caller(1)
	start := 0
	if ok {
		start = bytes.Index(stack, []byte(fmt.Sprintf("%s:%d", filename, fileLine)))
		stack = stack[start:length]
	}

	start = bytes.Index(stack, line) + 1
	stack = stack[start:]
	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}
	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")
	return stack
}

// GetFuncName 获取某个函数的名字
func GetFuncName(f any) string {
	return strings.Trim(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "-fm")
}

// GetInterfaceType 获取接口的类型
func GetInterfaceType[T any](t *T) reflect.Type {
	return reflect.TypeOf(t).Elem()
}

func InjectWrapFn(cemetery Cemetery, fn any) (*reflect.Value, error) {
	ft := reflect.TypeOf(fn)
	fv := reflect.ValueOf(fn)
	if ft.Kind() != reflect.Func {
		return nil, errors.New("fn must be a function")
	}

	in := ft.NumIn()
	if in > 1 {
		return nil, errors.New("fn only support one input parameter or no input parameter")
	}

	var args = make([]reflect.Value, 0)

	if in == 1 {
		if ft.In(0).Kind() != reflect.Struct {
			return nil, errors.New("fn input parameter must be a struct")
		}

		pt := ft.In(0)

		if pt.Name() != "" || pt.PkgPath() != "" {
			return nil, errors.New("fn input parameter must be a anonymous struct")
		}

		parameter := reflect.New(pt)

		goner := parameter.Interface()
		_, err := cemetery.ReviveOne(goner)
		if err != nil {
			return nil, errors.New("ReviveOne failed:" + err.Error())
		}
		args = append(args, parameter.Elem())
	}

	var outList []reflect.Type
	for i := 0; i < ft.NumOut(); i++ {
		outList = append(outList, ft.Out(i))
	}

	makeFunc := reflect.MakeFunc(reflect.FuncOf(nil, outList, false), func([]reflect.Value) (results []reflect.Value) {
		return fv.Call(args)
	})
	return &makeFunc, nil
}

func ExecuteInjectWrapFn(fn *reflect.Value) (results []any) {
	call := fn.Call([]reflect.Value{})

	for i := 0; i < len(call); i++ {
		results = append(results, call[i].Interface())
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
