package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
)

// RouterGroupName 路由分组名称
type RouterGroupName string

type Context struct {
	*gin.Context
}

// HandlerFunc `gone`框架的路由处理函数
type HandlerFunc func(*Context) (interface{}, error)

// IRoutes `gone`框架基于`gin`封装的路由，用于定义处理特定请求的函数
// 注入默认的路由使用Id: gone-gin-router (`gone.IdGoneGinRouter`)
// 给对象`inject`路由依赖：
// ```go
//
//	func NewDependOnIRoutes() *DependOnIRoutes {
//		return &DependOnIRoutes{}
//	}
//
//
//	type DependOnIRoutes struct {
//		TheRoutes gone.IRoutes `gone:"gone-gin-router"` //依赖注入系统路由
//	}
//
//	func (*DependOnIRoutes)DoSomething()  {
//		//对路由进行操作
//		//...
//	}
//
// ```
type IRoutes interface {
	// Use 在路由上应用`gin`中间件
	Use(...gin.HandlerFunc) IRoutes

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

// IRouter `gone`框架基于`gin`封装的"路由器"
// 注入默认的路由器使用Id: gone-gin-router (`gone.IdGoneGinRouter`)
type IRouter interface {
	// IRoutes 1. 组合了`gone.IRoutes`，可以定义路由
	IRoutes

	// GetGinRouter 2. 可以获取被封装的ginRouter对象，用于操作原始的gin路由
	GetGinRouter() gin.IRouter

	// Group 3.定义路由分组
	Group(string, ...gin.HandlerFunc) RouteGroup
}

// RouteGroup 路由分组
// 注入默认的路由分组使用Id: gone-gin-router (`gone.IdGoneGinRouter`)
type RouteGroup interface {
	IRouter
}

// Controller 控制器接口，由业务代码编码实现，用于挂载和处理路由
// 使用方式参考 [示例代码](https://gitlab.openviewtech.com/gone/gone-example/-/tree/master/gone-app)
type Controller interface {
	// Mount   路由挂载接口，改接口会在服务启动前被调用，该函数的实现通常情况应该返回`nil`
	Mount() MountError
}

// MountError `gin.Controller#Mount`返回的类型，用于识别 `gin.Controller` 的实现，避免被错误的调用到
type MountError error

// ## 错误处理
// 我们定义两种错误接口，3种具体的错误

// HandleProxyToGin 代理器，提供一个proxy函数将`gin.HandlerFunc`转成`gin.HandlerFunc`
// 注入`gin.HandleProxyToGin`使用Id：sys-gone-proxy (`gin.SystemGoneProxy`)
type HandleProxyToGin interface {
	Proxy(handler ...HandlerFunc) []gin.HandlerFunc
}

type jsonWriter interface {
	JSON(code int, obj any)
}

// Responser 响应处理器
// 注入默认的响应处理器使用Id: gone-gin-responser (`gone.IdGoneGinResponser`)
type Responser interface {
	gone.Goner
	Success(ctx jsonWriter, data interface{})
	Failed(ctx jsonWriter, err error)
}

type Close func()

// Server `gone`服务，可以代表一个`gone`应用
// 注入`gin.Server`使用Id：gone-gin (`gone.IdGoneGin`)
type Server interface {
	gone.Angel

	// Serve 启动http服务，返回的函数可以用于"服务优雅停机"
	Serve() (close Close)
}

// BusinessError
// 2. BusinessError，业务错误
// 业务错误是业务上的特殊情况，需要在不同的业务场景返回不同的数据类型；本质上不算错误，是为了便于业务编写做的一种抽象，
// 让同一个接口拥有在特殊情况返回不同业务代码和业务数据的能力
type BusinessError interface {
	gone.Error
	Data() interface{}
}
