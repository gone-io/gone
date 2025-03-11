package main

import "github.com/gone-io/gone/v2"

type Dep struct {
	gone.Flag
	Name string
}

type Component struct {
	gone.Flag
	dep *Dep        `gone:"*"` //依赖注入
	log gone.Logger `gone:"*"`
}

func (c *Component) Init() {
	c.log.Infof(c.dep.Name) //使用依赖
}

func main() {
	gone.
		NewApp().
		Load(&Dep{Name: "Component Dep"}).
		Load(&Component{}).
		Loads(func(l gone.Loader) error {
			_ = l.Load(&Component{})
			_ = l.Load(&Dep{Name: "Loads Dep"})
			return nil
		}).
		Run()
}
