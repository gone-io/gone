package main

import (
	"github.com/gone-io/gone"
	"use_default/service"
)

func main() {
	gone.Default.Run(func(i struct {
		iFood    service.IFood    `gone:"*"`
		iSession service.ISession `gone:"*"`
		iUser    service.IUser    `gone:"*"`
	}) {
		_ = i.iSession.Put("ok")
		_ = i.iFood.Create()
		hello := i.iUser.Hello()
		println(hello)
	})
}
