package gin

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"net/http"
	"reflect"
)

// NewGinResponser 新建系统默认的响应处理器
// 注入的ID为：gone-gin-responser (`gone.IdGoneGinResponser`)
func NewGinResponser() (gone.Goner, gone.GonerId) {
	return &responser{
		wrappedDataFunc: wrapFunc,
	}, gone.IdGoneGinResponser
}

type res[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
	Data T      `json:"data,omitempty"`
}

func wrapFunc(code int, msg string, data any) any {
	return &res[any]{Code: code, Msg: msg, Data: data}
}

type responser struct {
	gone.Flag
	logrus.Logger `gone:"gone-logger"`
	tracer        tracer.Tracer `gone:"gone-tracer"`

	wrappedDataFunc   WrappedDataFunc
	returnWrappedData bool `gone:"config,server.return.wrapped-data,default=true"`
}

func (r *responser) SetWrappedDataFunc(wrappedDataFunc WrappedDataFunc) {
	r.wrappedDataFunc = wrappedDataFunc
}

func noneWrappedData(ctx *gin.Context, data any) {
	t := reflect.TypeOf(data)
	switch t.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		ctx.JSON(http.StatusOK, data)

	case reflect.Pointer:
		switch t.Elem().Kind() {
		case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
			ctx.JSON(http.StatusOK, data)
		default:
			ctx.String(http.StatusOK, fmt.Sprintf("%s", data))
		}
	default:
		ctx.String(http.StatusOK, fmt.Sprintf("%s", data))
	}
}

func (r *responser) Success(ctx *gin.Context, data any) {
	if !r.returnWrappedData {
		if data == nil {
			return
		}
		noneWrappedData(ctx, data)
		return
	}

	if bErr, ok := data.(BusinessError); ok {
		ctx.JSON(http.StatusOK, wrapFunc(bErr.Code(), bErr.Msg(), bErr.Data()))
		return
	}

	ctx.JSON(http.StatusOK, wrapFunc(0, "", data))
}

func (r *responser) Failed(ctx *gin.Context, oErr error) {
	err := ToError(oErr)
	if !r.returnWrappedData {
		if err == nil {
			return
		}
		var iErr gone.InnerError
		if errors.As(err, &iErr) {
			ctx.String(http.StatusInternalServerError, iErr.Msg())
			r.tracer.Go(func() {
				r.Errorf("inner Error: %s(code=%d)\n%s", iErr.Msg(), iErr.Code(), iErr.Stack())
			})
			return
		}

		noneWrappedData(ctx, err)
		return
	}

	var bErr BusinessError
	if errors.As(err, &bErr) {
		ctx.JSON(http.StatusOK, wrapFunc(bErr.Code(), bErr.Msg(), bErr.Data()))
		return
	}

	var iErr gone.InnerError
	if errors.As(err, &iErr) {
		ctx.JSON(http.StatusInternalServerError, wrapFunc(iErr.Code(), iErr.Error(), error(nil)))
		r.tracer.Go(func() {
			r.Errorf("inner Error: %s(code=%d)\n%s", iErr.Msg(), iErr.Code(), iErr.Stack())
		})
		return
	}
	ctx.JSON(http.StatusBadRequest, wrapFunc(err.Code(), err.Msg(), error(nil)))
}
