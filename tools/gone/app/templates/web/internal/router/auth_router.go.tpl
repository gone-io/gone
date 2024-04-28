package router

import (
	"demo-structure/internal/middleware"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/gin"
)

const IdAuthRouter = "router-auth"

//go:gone
func NewAuth() (gone.Goner, gone.GonerId) {
	return &authRouter{}, IdAuthRouter
}

type authRouter struct {
	gone.Flag
	gin.IRouter
	root gin.IRouter `gone:"gone-gin-router"`

	auth *middleware.AuthorizeMiddleware `gone:"*"`
	pub  *middleware.PubMiddleware       `gone:"*"`
}

func (r *authRouter) AfterRevive() gone.AfterReviveError {
	r.IRouter = r.root.Group("/api", r.pub.Next, r.auth.Next)
	return nil
}
