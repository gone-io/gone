package main

import (
	"github.com/gone-io/gone/tools/gone/app"
	"github.com/gone-io/gone/tools/gone/mock"
	"github.com/gone-io/gone/tools/gone/priest"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func create() *cli.App {
	return &cli.App{
		Name:        "gone",
		Description: "generate gone code or generate gone app",
		Commands: []*cli.Command{
			priest.CreateCommand(),
			mock.CreateCommand(),
			app.CreateCommand(),
		},
	}
}

func run(args ...string) int {
	tool := create()
	err := tool.Run(args)
	if err != nil {
		log.Println(err.Error())
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args...))
}
