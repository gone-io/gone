package controller

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/gin"
)

//go:gone
func NewUserController() gone.Goner {
	return &user{}
}

type user struct {
	gone.Flag
	authRouter gin.IRouter `gone:"router-auth"`
}

func (ctr *user) Mount() gin.MountError {
	ctr.authRouter.
		GET("/users/:id", ctr.getUserById).
		GET("/empty-test", ctr.getEmpty).
		GET("/ok", ctr.ok)
	return nil
}

type User struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (ctr *user) getUserById(context *gin.Context) (any, error) {
	id := context.Param("id")
	return &User{
		Id:   id,
		Name: "test",
	}, nil
}

func (ctr *user) getEmpty(*gin.Context) (any, error) {
	return nil, nil
}

func (ctr *user) ok(*gin.Context) (any, error) {
	return "ok", nil
}
