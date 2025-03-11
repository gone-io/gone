package main

import "github.com/gone-io/gone/v2"

type Business struct {
	gone.Flag
	structInjector gone.StructInjector `gone:"*"`
}

type Dep struct {
	gone.Flag
	Name string
}

func (b *Business) BusinessProcess() {
	type User struct {
		Dep *Dep `gone:"*"`
	}

	var user User

	err := b.structInjector.InjectStruct(&user)
	if err != nil {
		panic(err)
	}
	println("user.Dep.name->", user.Dep.Name)
}

func main() {
	gone.
		Load(&Business{}).
		Load(&Dep{Name: "dep"}).
		Run(func(b *Business) {
			b.BusinessProcess()
		})
}
