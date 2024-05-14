package controller

import (
	"fmt"
	"github.com/gone-io/gone"
	"net/http"
	"net/url"
	"web-mysql/internal/interface/domain"
	"web-mysql/internal/interface/service"
	"web-mysql/internal/pkg/utils"
)

//go:gone
func NewDemoController() gone.Goner {
	return &demoController{}
}

type demoController struct {
	gone.Flag
	demoSvc     service.IDemo `gone:"*"`
	gone.Logger `gone:"gone-logger"`

	authRouter gone.IRouter `gone:"router-auth"`
	pubRouter  gone.IRouter `gone:"router-pub"`
}

func (ctr *demoController) Mount() gone.GinMountError {

	//需要鉴权的路由分组
	ctr.
		authRouter.
		Group("/demo").
		GET("/show", ctr.showDemo)

	//不需要鉴权的路由分组
	ctr.
		pubRouter.
		Group("/demo2").
		GET("/show", ctr.showDemo).
		GET("/error", ctr.error).
		GET("/echo", ctr.echo)

	// http注入 演示组
	ctr.
		pubRouter.
		GET("/inject-query", ctr.injectQuery).
		GET("/inject/:key", ctr.injectUrlParam).
		POST("/inject-http-body", ctr.injectHttpBody).
		GET("/inject-http-struct", ctr.injectHttpStruct)

	// demo数据 user 的增删改查，挂载到authRouter只为方便演示
	ctr.
		pubRouter.
		Group("/users").
		POST("", func(in struct {
			req *domain.User `gone:"http,body"`
		}) error {
			return ctr.demoSvc.CreateUser(in.req)
		}).
		GET("", func() (any, error) {
			return ctr.demoSvc.ListUsers()
		}).
		GET("/page", func(in struct {
			query domain.PageQuery `gone:"http,query"`
		}) (any, error) {
			return ctr.demoSvc.PageUsers(in.query)
		}).
		GET("/:id", func(in struct {
			id int64 `gone:"http,param"`
		}) (any, error) {
			return ctr.demoSvc.GetUserById(in.id)
		}).
		PUT("/:id", func(in struct {
			id  int64        `gone:"http,param"`
			req *domain.User `gone:"http,body"`
		}) error {
			return ctr.demoSvc.UpdateUserById(in.id, in.req)
		}).
		DELETE("/:id", func(in struct {
			id int64 `gone:"http,param"`
		}) error {
			return ctr.demoSvc.DeleteUser(in.id)
		})

	return nil
}

type QueryReq struct {
	Key string `form:"key"`
}

// http注入query参数
func (ctr *demoController) injectQuery(in struct {
	key      string  `gone:"http,query"`     //参数类型为string，标签中没有指定参数名时取属性名
	keyFloat float64 `gone:"http,query=key"` //参数类型为float64，标签中指定了参数名为key
	keyArr   []int   `gone:"http,query=key"` //参数类型为[]int，标签中指定了参数名为key

	//参数类型为*QueryReq；
	//类型为一个结构体或者结构体指针时，会尝试将query中的参数值注入到结构体中；结构体的属性标注和gin框架的标签标注一致
	query *QueryReq `gone:"http,query"`
}) (string, error) {
	query := fmt.Sprintf("key:%s;keyFloat:%f;keyArr:%v", in.key, in.keyFloat, in.keyArr)
	ctr.Infof("query=> %s", query)
	ctr.Infof("query=> %v", in.query)
	return query, nil
}

func (ctr *demoController) injectUrlParam(in struct {
	key    string `gone:"http,param"`
	keyInt int    `gone:"http,param=key"`
}) (any, error) {
	ctr.Infof("key: %s;keyInt: %d", in.key, in.keyInt)
	return in.key, nil
}

func (ctr *demoController) injectHttpBody(in struct {
	req *Req `gone:"http,body"`
}) (any, error) {
	ctr.Infof("req.Echo: %s", in.req.Echo)
	return in.req, nil
}

func (ctr *demoController) injectHttpStruct(in struct {
	ctx    *gone.Context `gone:"http"`
	req    *http.Request `gone:"http"`
	url    *url.URL      `gone:"http"`
	header http.Header   `gone:"http"`
}) (any, error) {
	ctr.Infof("remote ip:%s", in.ctx.RemoteIP())
	ctr.Infof("method: %s", in.req.Method)
	ctr.Infof("url path:%s", in.url.Path)
	ctr.Infof("header: %v", in.header)

	return "ok", nil
}

type Req struct {
	Echo string `form:"echo" json:"echo" xml:"echo"`
}

func (ctr *demoController) showDemo(ctx *gone.Context) (any, error) {
	return ctr.demoSvc.Show()
}

func (ctr *demoController) error(ctx *gone.Context) (any, error) {
	return ctr.demoSvc.Error()
}

func (ctr *demoController) echo(ctx *gone.Context) (any, error) {
	type Req struct {
		Echo string `form:"echo"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return nil, gone.NewParameterError(err.Error(), utils.ParameterParseError)
	}
	return ctr.demoSvc.Echo(req.Echo)
}
