package main

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner"
)

type controller struct {
	gone.Flag
	root gone.RouteGroup `gone:"*"`
}

func (c *controller) Mount() gone.GinMountError {
	c.root.GET("/test", func() string {
		return "hello world"
	})
	return nil
}

func init() {
	gone.
		Loads(goner.GinLoad).
		Load(&controller{})
}

func main() {
	gone.Serve()
}
