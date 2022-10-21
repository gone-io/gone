package router

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/gin"
)

const IdRouterAuth = "router-auth"

//go:gone
func NewAuthRouter() (gone.Goner, gone.GonerId) {
	return &authRouter{}, IdRouterAuth
}

type authRouter struct {
	gone.Flag
	gin.IRouter
	origin gin.IRouter `gone:"gone-gin-router"`
}

func (r *authRouter) AfterRevive() gone.AfterReviveError {
	r.IRouter = r.origin.Group("/api")
	return nil
}
