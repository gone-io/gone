package controller

import (
	"demo-structure/internal/interface/service"
	"demo-structure/internal/pkg/utils"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/gin"
)

//go:gone
func NewDemoController() gone.Goner {
	return &demoController{}
}

type demoController struct {
	gone.Flag
	demoSvc service.IDemo `gone:"*"`

	authRouter gin.IRouter `gone:"router-auth"`
	pubRouter  gin.IRouter `gone:"router-pub"`
}

func (ctr *demoController) Mount() gin.MountError {

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

	return nil
}

func (ctr *demoController) showDemo(ctx *gin.Context) (any, error) {
	return ctr.demoSvc.Show()
}

func (ctr *demoController) error(ctx *gin.Context) (any, error) {
	return ctr.demoSvc.Error()
}

func (ctr *demoController) echo(ctx *gin.Context) (any, error) {
	type Req struct {
		Echo string `form:"echo"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return nil, gin.NewParameterError(err.Error(), utils.ParameterParseError)
	}
	return ctr.demoSvc.Echo(req.Echo)
}
