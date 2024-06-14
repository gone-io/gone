package gone

import "github.com/gin-gonic/gin"

// Context is a wrapper of gin.Context
type Context struct {
	*gin.Context
}

type ResponseWriter = gin.ResponseWriter

type HandlerFunc any

type IRoutes interface {
	Use(...HandlerFunc) IRoutes

	Handle(string, string, ...HandlerFunc) IRoutes
	Any(string, ...HandlerFunc) IRoutes
	GET(string, ...HandlerFunc) IRoutes
	POST(string, ...HandlerFunc) IRoutes
	DELETE(string, ...HandlerFunc) IRoutes
	PATCH(string, ...HandlerFunc) IRoutes
	PUT(string, ...HandlerFunc) IRoutes
	OPTIONS(string, ...HandlerFunc) IRoutes
	HEAD(string, ...HandlerFunc) IRoutes
}

type IRouter interface {
	IRoutes

	GetGinRouter() gin.IRouter

	Group(string, ...HandlerFunc) RouteGroup

	LoadHTMLGlob(pattern string)
}

// RouteGroup route group, which is a wrapper of gin.RouterGroup, and can be injected for mount router.
type RouteGroup interface {
	IRouter
}

type GinMiddleware interface {
	Process(ctx *Context) error
}

type GinMountError error
