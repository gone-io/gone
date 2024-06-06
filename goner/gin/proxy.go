package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"reflect"
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

var ctxPtr *gin.Context
var ctxPointType = reflect.TypeOf(ctxPtr)
var ctxType = ctxPointType.Elem()

var goneContextPtr *gone.Context
var goneContextPointType = reflect.TypeOf(goneContextPtr)
var goneContextType = goneContextPointType.Elem()

type placeholder struct {
	Type reflect.Type
}

type BindStructFuncAndType struct {
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
		}
	default:
		return p.buildProxyFn(x, funcName, last)
	}
}

func (p *proxy) buildProxyFn(x HandlerFunc, funcName string, last bool) gin.HandlerFunc {
	m := make(map[int]*BindStructFuncAndType)
	args, err := p.cemetery.InjectFuncParameters(
		x,
		func(pt reflect.Type, i int) any {
			switch pt {
			case ctxPointType, ctxType, goneContextPointType, goneContextType:
				return &placeholder{
					Type: pt,
				}
			}
			p.injector.StartCollectBindFuncs()
			return nil
		},
		func(pt reflect.Type, i int, obj any) {
			m[i] = &BindStructFuncAndType{
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
		defer gone.TimeStat(funcName)()

		parameters := make([]reflect.Value, 0, len(args))
		for i, arg := range args {
			if holder, ok := arg.(*placeholder); ok {
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
				continue
			}

			if f, ok := m[i]; ok {
				parameter, err := f.Fn(context, arg, f.Type)
				if err != nil {
					p.responser.Failed(context, err)
					return
				}
				parameters = append(parameters, parameter)
				continue
			}
			parameters = append(parameters, reflect.ValueOf(arg))
		}

		//call the func x
		values := fv.Call(parameters)

		var results []any
		for i := 0; i < len(values); i++ {
			arg := values[i]
			if arg.Type() == ctxPointType && !arg.IsNil() {
				results = append(results, nil)
			} else {
				results = append(results, arg.Interface())
			}
		}
		p.responser.ProcessResults(context, context.Writer, last, funcName, results...)
	}
}
