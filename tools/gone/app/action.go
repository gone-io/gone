package app

import (
	"github.com/urfave/cli/v2"
	"path"
)

func Action(c *cli.Context) error {
	return doAction(c.String("template"), c.String("mod"), c.Args().Get(0))
}

func doAction(template, modName, appName string) error {
	tplName, app, mod, err := paramsProcess(template, appName, modName)
	if err != nil {
		return err
	}
	return copyAndReplace(f,
		path.Join(".", tplName),
		app,
		map[string]string{
			"${{appName}}":   app,
			"demo-structure": mod,
		},
	)
}
