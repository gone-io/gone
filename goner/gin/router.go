package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"net/http"
)

// NewGinRouter 用于创建系统根路由
func NewGinRouter() (gone.Goner, gone.GonerId, gone.GonerOption) {
	return &router{id: incr}, gone.IdGoneGinRouter, gone.IsDefault(true)
}

var incr = 0

type router struct {
	gone.Flag

	id int
	r  gin.IRouter
	*gin.Engine

	htmlTpl string `gone:"config,server.html-tpl-pattern"`
	mode    string `gone:"config,server.mode,default=release"`

	HandleProxyToGin `gone:"gone-gin-proxy"`
}

func (r *router) AfterRevive() gone.AfterReviveError {
	gin.SetMode(r.mode)
	r.Engine = gin.New()

	if r.htmlTpl != "" {
		r.Engine.LoadHTMLGlob(r.htmlTpl)
	}
	return nil
}

func (r *router) GetGinRouter() gin.IRouter {
	return r.Engine
}

func (r *router) getR() gin.IRouter {
	if r.r == nil {
		r.r = r.Engine
	}
	return r.r
}

func (r *router) Use(middleware ...HandlerFunc) IRoutes {
	r.getR().Use(r.ProxyForMiddleware(middleware...)...)
	return r
}

func (r *router) Group(relativePath string, handlers ...HandlerFunc) RouteGroup {
	incr++
	return &router{
		id:               incr,
		r:                r.getR().Group(relativePath, r.ProxyForMiddleware(handlers...)...),
		Engine:           r.Engine,
		HandleProxyToGin: r.HandleProxyToGin,
	}
}

func (r *router) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) IRoutes {
	r.getR().Handle(httpMethod, relativePath, r.Proxy(handlers...)...)
	return r
}
func (r *router) Any(relativePath string, handlers ...HandlerFunc) IRoutes {
	r.getR().Handle(http.MethodGet, relativePath, r.Proxy(handlers...)...)
	r.getR().Handle(http.MethodPost, relativePath, r.Proxy(handlers...)...)
	r.getR().Handle(http.MethodPut, relativePath, r.Proxy(handlers...)...)
	r.getR().Handle(http.MethodPatch, relativePath, r.Proxy(handlers...)...)
	r.getR().Handle(http.MethodHead, relativePath, r.Proxy(handlers...)...)
	r.getR().Handle(http.MethodOptions, relativePath, r.Proxy(handlers...)...)
	r.getR().Handle(http.MethodDelete, relativePath, r.Proxy(handlers...)...)
	r.getR().Handle(http.MethodConnect, relativePath, r.Proxy(handlers...)...)
	r.getR().Handle(http.MethodTrace, relativePath, r.Proxy(handlers...)...)
	return r
}
func (r *router) GET(relativePath string, handlers ...HandlerFunc) IRoutes {
	r.getR().Handle(http.MethodGet, relativePath, r.Proxy(handlers...)...)
	return r
}
func (r *router) POST(relativePath string, handlers ...HandlerFunc) IRoutes {
	r.getR().Handle(http.MethodPost, relativePath, r.Proxy(handlers...)...)
	return r
}
func (r *router) DELETE(relativePath string, handlers ...HandlerFunc) IRoutes {
	r.getR().Handle(http.MethodDelete, relativePath, r.Proxy(handlers...)...)
	return r
}
func (r *router) PATCH(relativePath string, handlers ...HandlerFunc) IRoutes {
	r.getR().Handle(http.MethodPatch, relativePath, r.Proxy(handlers...)...)
	return r
}
func (r *router) PUT(relativePath string, handlers ...HandlerFunc) IRoutes {
	r.getR().Handle(http.MethodPut, relativePath, r.Proxy(handlers...)...)
	return r
}
func (r *router) OPTIONS(relativePath string, handlers ...HandlerFunc) IRoutes {
	r.getR().Handle(http.MethodOptions, relativePath, r.Proxy(handlers...)...)
	return r
}
func (r *router) HEAD(relativePath string, handlers ...HandlerFunc) IRoutes {
	r.getR().Handle(http.MethodHead, relativePath, r.Proxy(handlers...)...)
	return r
}
