package gone

import (
	"net/http"
)

type BusinessError interface {
	Error
	Data() any
}

//三种错误：
//内部错误，系统内部错误，只能通过系统升级来修复
//参数错误，输入性错误，错误是由于输入信息导致的问题，需要调整输入
//业务错误，由于内部或外部信息导向的不同业务结果

//func NewInnerError(msg string, code int) Error {
//	return gone.NewInnerError(code, msg)
//}

// NewParameterError 新建`参数错误`
func NewParameterError(msg string, ext ...int) Error {
	var code = http.StatusBadRequest
	if len(ext) > 0 {
		code = ext[0]
	}
	return NewError(code, msg)
}

// NewBusinessError 新建`业务错误`
func NewBusinessError(msg string, ext ...any) BusinessError {
	var code = 0
	var data any = nil
	if len(ext) > 0 {
		i, ok := ext[0].(int)
		if ok {
			code = i
		}
	}
	if len(ext) > 1 {
		data = ext[1]
	}

	return &BError{err: NewError(code, msg), data: data}
}

// ToError 将 golang 提供的 error 转为一个 `gone.Error`
func ToError(err error) Error {
	if err == nil {
		return nil
	}
	if iErr, ok := err.(Error); ok {
		return iErr
	}
	return NewInnerError(err.Error(), http.StatusInternalServerError)
}

// BError 业务错误
type BError struct {
	err  Error
	data any
}

func (e *BError) Msg() string {
	return e.err.Msg()
}
func (e *BError) Code() int {
	return e.err.Code()
}
func (e *BError) Error() string {
	return e.err.Error()
}

func (e *BError) Data() any {
	return e.data
}
