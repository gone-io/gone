package main

import (
	"github.com/gone-io/gone/tools/gone/app"
	"github.com/gone-io/gone/tools/gone/mock"
	"github.com/gone-io/gone/tools/gone/priest"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	App := &cli.App{
		Name:           "gone",
		Description:    "generate gone code or generate gone app",
		DefaultCommand: "priest",
		Commands: []*cli.Command{
			priest.Command,
			mock.Command,
			app.Command,
		},
	}

	err := App.Run(os.Args)
	if err != nil {
		log.Fatalln("err:", err)
	}
}
