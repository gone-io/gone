package main

import (
	"github.com/gone-io/gone"
	server "github.com/gone-io/gone/example/http-server"
)

func main() {
	gone.Run(server.Priest)
}
