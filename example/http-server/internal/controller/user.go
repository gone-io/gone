package controller

import (
	gin2 "github.com/gin-gonic/gin"
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
	ctr.authRouter.GET("/users/:id", ctr.getUserById)
	return nil
}

type User struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (ctr *user) getUserById(context *gin2.Context) (interface{}, error) {
	id := context.Param("id")
	return &User{
		Id:   id,
		Name: "test",
	}, nil
}
