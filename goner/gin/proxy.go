package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
)

// NewGinProxy 新建代理器
func NewGinProxy() (gone.Goner, gone.GonerId) {
	return &proxy{}, gone.IdGoneGinProxy
}

type proxy struct {
	gone.Flag
	logrus.Logger `gone:"gone-logger"`
	handler       Responser `gone:"gone-gin-responser"`
}

func (p *proxy) Proxy(handlers ...HandlerFunc) (arr []gin.HandlerFunc) {
	count := len(handlers)
	for i := 0; i < count-2; i++ {
		arr = append(arr, p.proxyOne(handlers[i], false))
	}
	arr = append(arr, p.proxyOne(handlers[count-1], true))
	return arr
}

func (p *proxy) ProxyForMiddleware(handlers ...HandlerFunc) (arr []gin.HandlerFunc) {
	count := len(handlers)
	for i := 0; i < count-1; i++ {
		arr = append(arr, p.proxyOne(handlers[i], false))
	}
	return arr
}

func (p *proxy) proxyOne(handle HandlerFunc, last bool) gin.HandlerFunc {
	return func(context *gin.Context) {
		data, err := handle(&Context{Context: context})
		if err != nil {
			p.handler.Failed(context, err)
		}

		if data == nil {
			if !context.Writer.Written() && last {
				p.handler.Success(context, data)
			}
		} else {
			if context.Writer.Written() {
				p.Warnf("content had been written，check fn(%s)，maybe shouldn't return data", gone.FuncName(handle))
				return
			}
			p.handler.Success(context, data)
		}
	}
}
