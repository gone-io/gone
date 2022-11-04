package gin

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"net/http"
)

// NewGinResponser 新建系统默认的响应处理器
// 注入的ID为：gone-gin-responser (`gone.IdGoneGinResponser`)
func NewGinResponser() (gone.Goner, gone.GonerId) {
	return &responser{}, gone.IdGoneGinResponser
}

type res[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
	Data T      `json:"data,omitempty"`
}

func newRes[T any](code int, msg string, data T) *res[T] {
	return &res[T]{Code: code, Msg: msg, Data: data}
}

type responser struct {
	gone.Flag
	logrus.Logger `gone:"gone-logger"`
	tracer        tracer.Tracer `gone:"gone-tracer"`
}

func (r *responser) Success(ctx jsonWriter, data any) {
	bErr, ok := data.(BusinessError)
	if ok {
		ctx.JSON(http.StatusOK, newRes(bErr.Code(), bErr.Msg(), bErr.Data()))
		return
	}
	ctx.JSON(http.StatusOK, newRes(0, "", data))
}

func (r *responser) Failed(ctx jsonWriter, oErr error) {
	err := ToError(oErr)
	bErr, ok := err.(BusinessError)
	if ok {
		ctx.JSON(http.StatusOK, newRes(bErr.Code(), bErr.Msg(), bErr.Data()))
		return
	}

	iErr, ok := err.(gone.InnerError)
	if ok {
		ctx.JSON(http.StatusInternalServerError, newRes(iErr.Code(), iErr.Error(), error(nil)))
		r.tracer.Go(func() {
			if ok {
				r.Errorf("inner Error: %s(code=%d)\n%s", iErr.Msg(), iErr.Code(), iErr.Stack())
			}
		})
		return
	}
	ctx.JSON(http.StatusBadRequest, newRes(err.Code(), err.Msg(), error(nil)))
}
