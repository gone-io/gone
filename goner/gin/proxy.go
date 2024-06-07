package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"reflect"
	"time"
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
	stat        bool          `gone:"config,server.proxy.stat,default=false"`
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

var ctxPtr *gin.Context
var ctxPointType = reflect.TypeOf(ctxPtr)
var ctxType = ctxPointType.Elem()

var goneContextPtr *gone.Context
var goneContextPointType = reflect.TypeOf(goneContextPtr)
var goneContextType = goneContextPointType.Elem()

type placeholder struct {
	Type reflect.Type
}

var placeholderType = reflect.TypeOf(placeholder{})

type bindStructFuncAndType struct {
	Fn   BindStructFunc
	Type reflect.Type
}

func (p *proxy) proxyOne(x HandlerFunc, last bool) gin.HandlerFunc {
	funcName := gone.GetFuncName(x)

	switch x.(type) {
	case func(*Context) (any, error):
		f := x.(func(*Context) (any, error))
		return func(context *gin.Context) {
			data, err := f(&Context{Context: context})
			p.responser.ProcessResults(context, context.Writer, last, funcName, data, err)
		}
	case func(*Context) error:
		f := x.(func(*Context) error)
		return func(context *gin.Context) {
			err := f(&Context{Context: context})
			p.responser.ProcessResults(context, context.Writer, last, funcName, err)
		}
	case func(*Context):
		f := x.(func(*Context))
		return func(context *gin.Context) {
			f(&Context{Context: context})
			p.responser.ProcessResults(context, context.Writer, last, funcName)
		}
	case func(ctx *gin.Context):
		return x.(func(ctx *gin.Context))

	case func(ctx *gin.Context) (any, error):
		f := x.(func(ctx *gin.Context) (any, error))
		return func(context *gin.Context) {
			data, err := f(context)
			p.responser.ProcessResults(context, context.Writer, last, funcName, data, err)
		}
	case func(ctx *gin.Context) error:
		f := x.(func(ctx *gin.Context) error)
		return func(context *gin.Context) {
			err := f(context)
			p.responser.ProcessResults(context, context.Writer, last, funcName, err)
		}
	case func():
		f := x.(func())
		return func(context *gin.Context) {
			f()
			p.responser.ProcessResults(context, context.Writer, last, funcName)
		}
	case func() (any, error):
		f := x.(func() (any, error))
		return func(context *gin.Context) {
			data, err := f()
			p.responser.ProcessResults(context, context.Writer, last, funcName, data, err)
		}
	case func() error:
		f := x.(func() error)
		return func(context *gin.Context) {
			err := f()
			p.responser.ProcessResults(context, context.Writer, last, funcName, err)
		}
	default:
		return p.buildProxyFn(x, funcName, last)
	}
}

func (p *proxy) buildProxyFn(x HandlerFunc, funcName string, last bool) gin.HandlerFunc {
	m := make(map[int]*bindStructFuncAndType)
	args, err := p.cemetery.InjectFuncParameters(
		x,
		func(pt reflect.Type, i int) any {
			switch pt {
			case ctxPointType, ctxType, goneContextPointType, goneContextType:
				return placeholder{
					Type: pt,
				}
			}
			p.injector.StartBindFuncs()
			return nil
		},
		func(pt reflect.Type, i int) {
			m[i] = &bindStructFuncAndType{
				Fn:   p.injector.BindFuncs(),
				Type: pt,
			}
		},
	)

	if err != nil {
		panic(err)
	}

	fv := reflect.ValueOf(x)
	return func(context *gin.Context) {
		if p.stat {
			defer gone.TimeStat(funcName+"-inject-proxy", time.Now(), p.Infof)
		}

		parameters := make([]reflect.Value, 0, len(args))
		for i, arg := range args {
			switch arg.Type() {
			case placeholderType:
				holder := arg.Interface().(placeholder)
				switch holder.Type {
				case ctxPointType:
					parameters = append(parameters, reflect.ValueOf(context))
				case ctxType:
					parameters = append(parameters, reflect.ValueOf(context).Elem())
				case goneContextPointType:
					parameters = append(parameters, reflect.ValueOf(&Context{Context: context}))
				case goneContextType:
					parameters = append(parameters, reflect.ValueOf(Context{Context: context}))
				}
			default:
				if f, ok := m[i]; ok {
					parameter, err := f.Fn(context, arg)
					if err != nil {
						p.responser.Failed(context, err)
						return
					}
					parameters = append(parameters, parameter)
				} else {
					parameters = append(parameters, arg)
				}
			}
		}

		//call the func x
		values := fv.Call(parameters)

		var results []any
		for i := 0; i < len(values); i++ {
			arg := values[i]

			if arg.Kind() == reflect.Interface {
				elem := arg.Elem()
				switch elem.Kind() {
				case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
					if elem.IsNil() {
						results = append(results, nil)
						continue
					}
				default:
				}
			}
			results = append(results, arg.Interface())
		}
		p.responser.ProcessResults(context, context.Writer, last, funcName, results...)
	}
}
