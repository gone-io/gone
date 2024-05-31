package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
)

// NewGinProxy 新建代理器
func NewGinProxy() (gone.Goner, gone.GonerId, gone.GonerOption) {
	return &proxy{}, gone.IdGoneGinProxy, gone.IsDefault(true)
}

type proxy struct {
	gone.Flag
	gone.Logger `gone:"*"`
	cemetery    gone.Cemetery `gone:"*"`
	responser   Responser     `gone:"*"`
	tracer      gone.Tracer   `gone:"*"`
	injector    HttInjector   `gone:"*"`
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
			p.responser.ProcessResults(context, context.Writer, last, gone.GetFuncName(x), data, err)
		}

	case func(*Context) error:
		return func(context *gin.Context) {
			err := x.(func(*Context) error)(&Context{Context: context})
			p.responser.ProcessResults(context, context.Writer, last, gone.GetFuncName(x), err)
		}
	case func(*Context):
		return func(context *gin.Context) {
			x.(func(*Context))(&Context{Context: context})
		}
	default:
		p.injector.StartCollectBindFuncs()
		fn, err := gone.InjectWrapFn(p.cemetery, x)
		if err != nil {
			panic(err)
		}
		funcs := p.injector.CollectBindFuncs()

		return func(context *gin.Context) {
			for _, f := range funcs {
				err := f(context)
				if err != nil {
					p.responser.Failed(context, err)
					return
				}
			}

			results := gone.ExecuteInjectWrapFn(fn)
			p.responser.ProcessResults(context, context.Writer, last, gone.GetFuncName(x), results...)
		}
	}
}
