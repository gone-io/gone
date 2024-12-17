package gin

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"io"
	"net/http"
	"reflect"
)

func NewGinResponser() gone.Goner {
	return &responser{
		wrappedDataFunc: wrapFunc,
	}
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
	gone.Logger `gone:"gone-logger"`

	wrappedDataFunc   WrappedDataFunc
	returnWrappedData bool `gone:"config,server.return.wrapped-data,default=true"`
}

func (r *responser) GonerName() string {
	return IdGoneGinResponser
}

func (r *responser) SetWrappedDataFunc(wrappedDataFunc WrappedDataFunc) {
	r.wrappedDataFunc = wrappedDataFunc
}

func noneWrappedData(ctx XContext, data any, status int) {
	if data == nil {
		ctx.String(status, "")
		return
	}

	if err, ok := data.(error); ok {
		ctx.String(status, err.Error())
		return
	}

	t := reflect.TypeOf(data)
	switch t.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		ctx.JSON(status, data)

	case reflect.Pointer:
		switch t.Elem().Kind() {
		case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
			ctx.JSON(status, data)
		default:
			ctx.String(status, fmt.Sprintf("%v", reflect.ValueOf(data).Elem().Interface()))
		}
	default:
		ctx.String(status, fmt.Sprintf("%v", data))
	}
}

func (r *responser) Success(ctx XContext, data any) {
	if !r.returnWrappedData {
		noneWrappedData(ctx, data, http.StatusOK)
		return
	}

	if bErr, ok := data.(BusinessError); ok {
		ctx.JSON(http.StatusOK, wrapFunc(bErr.Code(), bErr.Msg(), bErr.Data()))
		return
	}

	ctx.JSON(http.StatusOK, wrapFunc(0, "", data))
}

func (r *responser) Failed(ctx XContext, oErr error) {
	err := ToError(oErr)
	if !r.returnWrappedData {
		var iErr gone.InnerError
		if err == nil {
			noneWrappedData(ctx, nil, http.StatusBadRequest)
			return
		}
		if errors.As(err, &iErr) {
			ctx.String(http.StatusInternalServerError, iErr.Msg())
			r.Errorf("inner Error: %s(code=%d)\n%s", iErr.Msg(), iErr.Code(), iErr.Stack())
			return
		}
		noneWrappedData(ctx, err, err.GetStatusCode())
		return
	}

	if oErr == nil {
		ctx.JSON(http.StatusBadRequest, wrapFunc(0, "", nil))
		return
	}

	var bErr BusinessError
	if errors.As(err, &bErr) {
		ctx.JSON(bErr.GetStatusCode(), wrapFunc(bErr.Code(), bErr.Msg(), bErr.Data()))
		return
	}

	var iErr gone.InnerError
	if errors.As(err, &iErr) {
		ctx.JSON(iErr.GetStatusCode(), wrapFunc(iErr.Code(), "Internal Server Error", nil))
		r.Errorf("inner Error: %s(code=%d)\n%s", iErr.Msg(), iErr.Code(), iErr.Stack())
		return
	}
	ctx.JSON(err.GetStatusCode(), wrapFunc(err.Code(), err.Msg(), error(nil)))
}

func (r *responser) ProcessResults(context XContext, writer gin.ResponseWriter, last bool, funcName string, results ...any) {
	for _, result := range results {
		if err, ok := result.(error); ok {
			r.Failed(context, err)
			context.Abort()
			return
		}
	}

	isNotEnd := false
	for _, result := range results {
		if result == nil {
			continue
		}

		if writer.Written() && result != nil {
			r.Warnf("content had been written，check fn(%s)，maybe shouldn't return data", funcName)
			return
		}

		switch result.(type) {
		case error:
			r.Failed(context, result.(error))
		case io.Reader:
			isNotEnd = true
			_, err := io.Copy(writer, result.(io.Reader))
			if err != nil {
				r.Warnf("copy data to writer failed, err: %v", err)
			}
		case chan any:
			isNotEnd = true
			r.dealChan(result.(chan any), writer)
		default:
			r.Success(context, result)
		}
	}

	if !writer.Written() && last && !isNotEnd {
		r.Success(context, nil)
	}
}

func (r *responser) dealChan(ch <-chan any, writer gin.ResponseWriter) {
	sse := NewSSE(writer)
	sse.Start()

	for {
		data, ok := <-ch

		if !ok {
			err := sse.End()
			if err != nil {
				r.Errorf("write 'end' error: %v", err)
			}
			return
		}
		var err error
		switch data.(type) {
		case error:
			err = sse.WriteError(ToError(data.(error)))
		default:
			err = sse.Write(data)
		}

		if err != nil {
			r.Errorf("write data error: %v", err)
			return
		}
	}
}
