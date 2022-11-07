package gone

import (
	"bytes"
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
