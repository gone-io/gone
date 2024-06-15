package router

import (
	"demo-structure/internal/middleware"
	"github.com/gone-io/gone"
)

const IdAuthRouter = "router-auth"

//go:gone
func NewAuth() (gone.Goner, gone.GonerId) {
	return &authRouter{}, IdAuthRouter
}

type authRouter struct {
	gone.Flag
	gone.IRouter
	root gone.RouteGroup `gone:"*"`

	auth *middleware.AuthorizeMiddleware `gone:"*"`
	pub  *middleware.PubMiddleware       `gone:"*"`
}

func (r *authRouter) AfterRevive() gone.AfterReviveError {
	r.IRouter = r.root.Group("/api", r.pub.Next, r.auth.Next)
	return nil
}
