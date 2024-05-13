package gin

import (
	"github.com/gone-io/gone"
)

//三种错误：
//内部错误，系统内部错误，只能通过系统升级来修复
//参数错误，输入性错误，错误是由于输入信息导致的问题，需要调整输入
//业务错误，由于内部或外部信息导向的不同业务结果

func NewInnerError(msg string, code int) gone.Error {
	return gone.NewInnerError(msg, code)
}

// NewParameterError 新建`参数错误`
func NewParameterError(msg string, ext ...int) gone.Error {
	return gone.NewParameterError(msg, ext...)
}

// NewBusinessError 新建`业务错误`
func NewBusinessError(msg string, ext ...any) BusinessError {
	return gone.NewBusinessError(msg, ext...)
}

// ToError 将 golang 提供的 error 转为一个 `gone.Error`
func ToError(err error) gone.Error {
	return gone.ToError(err)
}

// 业务错误
type bError = gone.BError
