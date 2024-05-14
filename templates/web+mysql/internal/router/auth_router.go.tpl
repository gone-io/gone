package router

import (
	"github.com/gone-io/gone"
	"web-mysql/internal/middleware"
)

const IdAuthRouter = "router-auth"

//go:gone
func NewAuth() (gone.Goner, gone.GonerId) {
	return &authRouter{}, IdAuthRouter
}

type authRouter struct {
	gone.Flag
	gone.IRouter
	root gone.IRouter `gone:"gone-gin-router"`

	auth *middleware.AuthorizeMiddleware `gone:"*"`
	pub  *middleware.PubMiddleware       `gone:"*"`
}

func (r *authRouter) AfterRevive() gone.AfterReviveError {
	r.IRouter = r.root.Group("/api", r.pub.Next, r.auth.Next)
	return nil
}
