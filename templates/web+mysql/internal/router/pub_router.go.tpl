package router

import (
	"github.com/gone-io/gone"
	"web-mysql/internal/middleware"
)

const IdRouterPub = "router-pub"

//go:gone
func NewPubRouter() (gone.Goner, gone.GonerId) {
	return &pubRouter{}, IdRouterPub
}

type pubRouter struct {
	gone.Flag
	gone.IRouter
	root gone.RouteGroup           `gone:"*"`
	pub  *middleware.PubMiddleware `gone:"*"`
}

func (r *pubRouter) AfterRevive() gone.AfterReviveError {
	r.IRouter = r.root.Group("/api", r.pub.Next)
	return nil
}
