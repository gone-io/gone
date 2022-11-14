package app

import "github.com/urfave/cli/v2"

//for generate a gone app

var Command = &cli.Command{
	Name:        "app",
	Usage:       "${appName} -t ${template}",
	Description: "generate a gone app",
}
