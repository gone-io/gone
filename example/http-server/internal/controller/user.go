package controller

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/gin"
	"github.com/gone-io/gone/goner/logrus"
)

//go:gone
func NewUserController() gone.Goner {
	return &user{}
}

type user struct {
	gone.Flag
	authRouter gin.IRouter `gone:"router-auth"`
	pub        gin.IRouter `gone:"gone-gin-router"`
}

func (ctr *user) Mount() gin.MountError {
	ctr.authRouter.
		GET("/users/:id", ctr.getUserById).
		GET("/empty-test", ctr.getEmpty).
		GET("/ok", ctr.ok)

	ctr.pub.
		GET("/test", func(in struct {
			page     string `gone:"http,query=page"`
			cookX    string `gone:"http,cookie=x"`
			headerY  string `gone:"http,header=y"`
			token    string `gone:"http,auth=Bearer"`
			formData string `gone:"http,form=data"`

			host    string            `gone:"http,host"`
			url     string            `gone:"http,url"`
			path    string            `gone:"http,path"`
			query   map[string]string `gone:"http,query"`
			data    string            `gone:"http,body"`
			context *gin.Context      `gone:"http,context"`

			log logrus.Logger `gone:"gone-logger"`
		}) string {

			fmt.Printf("%v", in)

			return "ok"
		})
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
