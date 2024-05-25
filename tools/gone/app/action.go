package app

import (
	"github.com/urfave/cli/v2"
	"path"
)

func CreateCommand() *cli.Command {
	return &cli.Command{
		Name:        "create",
		Usage:       "[-t ${template} [-m ${modName}]] ${appName}",
		Description: "create a gone app",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "template",
				Aliases: []string{"t"},
				Value:   "web",
				Usage:   "template: only support web、web+mysql, more will be supported in the future",
			},

			&cli.StringFlag{
				Name:    "mod",
				Aliases: []string{"m"},
				Usage:   "modName",
			},
		},
		Action: action,
	}
}

func action(c *cli.Context) error {
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
