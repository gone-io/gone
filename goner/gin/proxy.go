package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"io"
)

// NewGinProxy 新建代理器
func NewGinProxy() (gone.Goner, gone.GonerId) {
	return &proxy{}, gone.IdGoneGinProxy
}

type proxy struct {
	gone.Flag
	logrus.Logger `gone:"gone-logger"`
	cemetery      gone.Cemetery `gone:"gone-cemetery"`
	handler       Responser     `gone:"gone-gin-responser"`
	tracer        tracer.Tracer `gone:"gone-tracer"`
}

func (p *proxy) Proxy(handlers ...HandlerFunc) (arr []gin.HandlerFunc) {
	count := len(handlers)
	for i := 0; i < count-1; i++ {
		arr = append(arr, p.proxyOne(handlers[i], false))
	}
	arr = append(arr, p.proxyOne(handlers[count-1], true))
	return arr
}

func (p *proxy) ProxyForMiddleware(handlers ...HandlerFunc) (arr []gin.HandlerFunc) {
	count := len(handlers)
	for i := 0; i < count; i++ {
		arr = append(arr, p.proxyOne(handlers[i], false))
	}
	return arr
}

func (p *proxy) proxyOne(x HandlerFunc, last bool) gin.HandlerFunc {
	switch x.(type) {
	case func(*Context) (any, error):
		return func(context *gin.Context) {
			data, err := x.(func(*Context) (any, error))(&Context{Context: context})
			if err != nil {
				p.handler.Failed(context, err)
				return
			}

			reader, ok := data.(io.Reader)
			if ok {
				_, err = io.Copy(context.Writer, reader)
				if err != nil {
					p.Warnf("copy data to writer failed, err: %v", err)
				}
				return
			}

			if context.Writer.Written() && data != nil {
				p.Warnf("content had been written，check fn(%s)，maybe shouldn't return data", gone.GetFuncName(x))
				return
			}

			if data != nil {
				p.handler.Success(context, data)
				return
			}

			if !context.Writer.Written() && last {
				p.handler.Success(context, data)
			}
		}

	case func(*Context) error:
		return func(context *gin.Context) {
			err := x.(func(*Context) error)(&Context{Context: context})
			if err != nil {
				p.handler.Failed(context, err)
			}
		}
	case func(*Context):
		return func(context *gin.Context) {
			x.(func(*Context))(&Context{Context: context})
		}

	default:
		return func(context *gin.Context) {
			fn, err := gone.InjectWrapFn(p.cemetery, x)
			if err != nil {
				p.Errorf("inject wrap fn failed, err: %v", err)
				p.handler.Failed(context, err)
				return
			}

			results := gone.ExecuteInjectWrapFn(fn)

			for _, result := range results {
				if result == nil {
					continue
				}
				if context.Writer.Written() && result != nil {
					p.Warnf("content had been written，check fn(%s)，maybe shouldn't return data", gone.GetFuncName(x))
					return
				}

				switch result.(type) {
				case error:
					p.handler.Failed(context, result.(error))
				case io.Reader:
					_, err = io.Copy(context.Writer, result.(io.Reader))
				case chan any:
					p.dealChan(result.(chan any), context)
				default:
					p.handler.Success(context, result)
				}
			}

			if !context.Writer.Written() && last {
				p.handler.Success(context, nil)
			}
		}
	}
}

func (p *proxy) dealChan(ch <-chan any, ctx *gin.Context) {
	sse := Sse{Context: ctx}
	sse.Start()

	for {
		select {
		case <-ctx.Request.Context().Done():
			return
		case data, ok := <-ch:
			if !ok {
				err := sse.End()
				if err != nil {
					p.Errorf("write 'end' error: %v", err)
				}
				return
			}
			var err error
			switch data.(type) {
			case gone.Error:
				err = sse.WriteError(data.(gone.Error))
			case error:
				err = sse.WriteError(ToError(data.(error)))
			default:
				err = sse.Write(data)
			}

			if err != nil {
				p.Errorf("write data error: %v", err)
				return
			}
		}
	}
}
