package gin

import (
	"github.com/gone-io/gone"
)

// NewInnerError 新建`内部错误`
var NewInnerError = gone.NewInnerError

// NewParameterError 新建`参数错误`
var NewParameterError = gone.NewParameterError

// NewBusinessError 新建`业务错误`
var NewBusinessError = gone.NewBusinessError

// ToError 转为错误
var ToError = gone.ToError

// 业务错误
type bError = gone.BError
