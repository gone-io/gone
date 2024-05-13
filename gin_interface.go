package gone

import "github.com/gin-gonic/gin"

// Context is a wrapper of gin.Context
type Context struct {
	*gin.Context
}

type ResponseWriter = gin.ResponseWriter

type HandlerFunc any

type IRoutes interface {
	// Use 在路由上应用`gin`中间件
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
	// IRoutes 1. 组合了`gone.IRoutes`，可以定义路由
	IRoutes

	// GetGinRouter 2. 可以获取被封装的ginRouter对象，用于操作原始的gin路由
	GetGinRouter() gin.IRouter

	// Group 3.定义路由分组
	Group(string, ...HandlerFunc) RouteGroup

	LoadHTMLGlob(pattern string)
}

// RouteGroup 路由分组
// 注入默认的路由分组使用Id: gone-gin-router (`gone.IdGoneGinRouter`)
type RouteGroup interface {
	IRouter
}

type GinMountError error
