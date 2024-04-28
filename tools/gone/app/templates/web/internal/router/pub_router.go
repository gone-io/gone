package router

import (
	"demo-structure/internal/middleware"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/gin"
)

const IdRouterPub = "router-pub"

//go:gone
func NewPubRouter() (gone.Goner, gone.GonerId) {
	return &pubRouter{}, IdRouterPub
}

type pubRouter struct {
	gone.Flag
	gin.IRouter
	root gin.IRouter               `gone:"gone-gin-router"`
	pub  *middleware.PubMiddleware `gone:"*"`
}

func (r *pubRouter) AfterRevive() gone.AfterReviveError {
	r.IRouter = r.root.Group("/api", r.pub.Next)
	return nil
}
