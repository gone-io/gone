package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"reflect"
)

//go:generate sh -c "mockgen -package=gin github.com/gin-gonic/gin ResponseWriter > gin_response_writer_mock_test.go"
//go:generate sh -c "mockgen -package=gin net Listener > net_listener_mock_test.go"
//go:generate sh -c "mockgen -package=gin -source=../../gin_interface.go |gone mock -o gone_gin_mock_test.go"
//go:generate sh -c "mockgen -package=gin -source=../../interface.go |gone mock -o gone_mock_test.go"
//go:generate sh -c "mockgen -package=gin -self_package=github.com/gone-io/gone/goner/gin -source=interface.go |gone mock -o mock_test.go"

// RouterGroupName Router group name
type RouterGroupName string

// Context The `gone` framework encapsulated context based on `gin`
// Deprecated use `gone.Context` instead
type Context = gone.Context

type OriginContent = gin.Context

// HandlerFunc The `gone` framework route handler function
type HandlerFunc = gone.HandlerFunc

// IRoutes Routes encapsulated by the `gone` framework based on `gin`, used to define functions that handle specific requests
// Inject default routes using Id: gone-gin-router (`gone.IdGoneGinRouter`)
// To inject route dependencies into an object:
// ```go
//
//	func NewDependOnIRoutes() *DependOnIRoutes {
//		return &DependOnIRoutes{}
//	}
//
//
//	type DependOnIRoutes struct {
//		TheRoutes gone.IRoutes `gone:"gone-gin-router"` // Dependency injection system routes
//	}
//
//	func (*DependOnIRoutes) DoSomething()  {
//		// Operate on the routes
//		//...
//	}
//
// ```
//
//	type IRoutes interface {
//		// Use Apply `gin` middleware on the route
//		Use(...HandlerFunc) IRoutes
//
//		Handle(string, string, ...HandlerFunc) IRoutes
//		Any(string, ...HandlerFunc) IRoutes
//		GET(string, ...HandlerFunc) IRoutes
//		POST(string, ...HandlerFunc) IRoutes
//		DELETE(string, ...HandlerFunc) IRoutes
//		PATCH(string, ...HandlerFunc) IRoutes
//		PUT(string, ...HandlerFunc) IRoutes
//		OPTIONS(string, ...HandlerFunc) IRoutes
//		HEAD(string, ...HandlerFunc) IRoutes
//	}
type IRoutes = gone.IRoutes

// IRouter The `gone` framework encapsulated "router" based on `gin`
// Inject default router using Id: gone-gin-router (`gone.IdGoneGinRouter`)
//
//	type IRouter interface {
//		// IRoutes 1. Composes `gone.IRoutes`, can define routes
//		IRoutes
//
//		// GetGinRouter 2. Can get the encapsulated ginRouter object, used to operate the original gin routes
//		GetGinRouter() gin.IRouter
//
//		// Group 3. Define route groups
//		Group(string, ...HandlerFunc) RouteGroup
//
//		LoadHTMLGlob(pattern string)
//	}
type IRouter = gone.IRouter

// RouteGroup Route group
// Inject default route group using Id: gone-gin-router (`gone.IdGoneGinRouter`)
//
//	type RouteGroup interface {
//		IRouter
//	}
type RouteGroup = gone.RouteGroup

// Controller interface, implemented by business code, used to mount and handle routes
// For usage reference [example code](https://gitlab.openviewtech.com/gone/gone-example/-/tree/master/gone-app)
type Controller interface {
	// Mount Route mount interface, this interface will be called before the service starts, the implementation of this function should usually return `nil`
	Mount() MountError
}

// MountError The type returned by `gin.Controller#Mount`, used to identify the implementation of `gin.Controller` to avoid being called incorrectly
type MountError = gone.GinMountError

// HandleProxyToGin Proxy, provides a proxy function to convert `gone.HandlerFunc` to `gin.HandlerFunc`
// Inject `gin.HandleProxyToGin` using Id: sys-gone-proxy (`gin.SystemGoneProxy`)
type HandleProxyToGin interface {
	Proxy(handler ...HandlerFunc) []gin.HandlerFunc
	ProxyForMiddleware(handlers ...HandlerFunc) (arr []gin.HandlerFunc)
}

type XContext interface {
	JSON(code int, obj any)
	String(code int, format string, values ...any)
}

type WrappedDataFunc func(code int, msg string, data any) any

// Responser Response handler
// Inject default response handler using Id: gone-gin-responser (`gone.IdGoneGinResponser`)
type Responser interface {
	Success(ctx XContext, data any)
	Failed(ctx XContext, err error)
	ProcessResults(context XContext, writer gin.ResponseWriter, last bool, funcName string, results ...any)
}

// BusinessError business error
// Business errors are special cases in business scenarios that need to return different data types in different business contexts; essentially not considered errors, but an abstraction to facilitate business writing,
// allowing the same interface to have the ability to return different business codes and business data in special cases
type BusinessError = gone.BusinessError

type BindFieldFunc func(context *gin.Context, structVale reflect.Value) error
type BindStructFunc func(*gin.Context, reflect.Value) (reflect.Value, error)

type HttInjector interface {
	StartBindFuncs()
	BindFuncs() BindStructFunc
}
