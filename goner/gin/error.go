package gin

import (
	"github.com/gone-io/gone"
	"net/http"
)

//三种错误：
//内部错误，系统内部错误，只能通过系统升级来修复
//参数错误，输入性错误，错误是由于输入信息导致的问题，需要调整输入
//业务错误，由于内部或外部信息导向的不同业务结果

func NewInnerError(code int, msg string) gone.Error {
	return gone.NewInnerError(code, msg)
}

// NewParameterError 新建`参数错误`
func NewParameterError(msg string, ext ...int) gone.Error {
	var code = http.StatusBadRequest
	if len(ext) > 0 {
		code = ext[0]
	}
	return gone.NewError(code, msg)
}

// NewBusinessError 新建`业务错误`
func NewBusinessError(msg string, ext ...interface{}) BusinessError {
	var code = 0
	var data interface{} = nil
	if len(ext) > 0 {
		i, ok := ext[0].(int)
		if ok {
			code = i
		}
	}
	if len(ext) > 1 {
		data = ext[1]
	}

	return &bError{err: gone.NewError(code, msg), data: data}
}

// ToError 将 golang 提供的 error 转为一个 `gone.Error`
func ToError(err error) gone.Error {
	iErr, ok := err.(gone.Error)
	if ok {
		return iErr
	}
	return NewInnerError(http.StatusInternalServerError, err.Error())
}

// 业务错误
type bError struct {
	err  gone.Error
	data interface{}
}

func (e *bError) Msg() string {
	return e.err.Msg()
}
func (e *bError) Code() int {
	return e.err.Code()
}
func (e *bError) Error() string {
	return e.err.Error()
}

func (e *bError) Data() interface{} {
	return e.data
}
