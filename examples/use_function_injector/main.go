package main

import "github.com/gone-io/gone/v2"

type Business struct {
	gone.Flag
	funcInjector gone.FuncInjector `gone:"*"`
}

type Dep struct {
	gone.Flag
	Name string
}

func (b *Business) BusinessProcess() {
	needInjectedFunc := func(dep *Dep) {
		println("dep.name->", dep.Name)
	}

	wrapFunc, err := b.funcInjector.InjectWrapFunc(needInjectedFunc, nil, nil)
	if err != nil {
		panic(err)
	}
	_ = wrapFunc()
}

func main() {
	gone.
		Load(&Business{}).
		Load(&Dep{Name: "dep"}).
		Run(func(b *Business) {
			b.BusinessProcess()
		})
}
