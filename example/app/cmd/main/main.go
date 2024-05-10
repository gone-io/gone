package main

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/example/app"
)

func main() {
	gone.Prepare(app.Priest).Serve()
}
