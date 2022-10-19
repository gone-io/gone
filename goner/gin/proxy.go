package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
)

// NewGinProxy 新建代理器
func NewGinProxy() (gone.Goner, gone.GonerId) {
	return &proxy{}, gone.IdGoneGinProxy
}

type proxy struct {
	gone.Flag
	handler Responser `gone:"gone-gin-responser"`
}

func (p *proxy) Proxy(handler ...HandlerFunc) (arr []gin.HandlerFunc) {
	for _, h := range handler {
		arr = append(arr, p.proxyOne(h))
	}
	return arr
}

func (p *proxy) proxyOne(handle HandlerFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		i, err := handle(&Context{Context: context})
		if err != nil {
			p.handler.Failed(context, err)
		}

		if context.Writer.Written() {
			return
		}
		p.handler.Success(context, i)
	}
}
