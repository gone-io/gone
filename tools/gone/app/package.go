package app

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/gone-io/gone/templates"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"path"
	"strings"
)

var appName, template, modName string

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:        "t",
		Value:       "web",
		Usage:       "template: only support web、web+mysql, more will be supported in the future",
		Destination: &template,
	},

	&cli.StringFlag{
		Name:        "m",
		Usage:       "modName",
		Destination: &modName,
	},
}

var f = templates.F

func copyAndReplace(f embed.FS, source, target string, replacement map[string]string) error {
	err := os.MkdirAll(target, 0766)
	if err != nil {
		return err
	}

	dir, err := f.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range dir {
		from := path.Join(source, entry.Name())
		to := path.Join(target, strings.TrimSuffix(entry.Name(), ".tpl"))

		if entry.IsDir() {
			err = copyAndReplace(
				f,
				from,
				to,
				replacement,
			)
			if err != nil {
				return err
			}
		} else {
			file, err := os.OpenFile(to, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return err
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Error(err)
				}
			}(file)

			data, err := f.ReadFile(from)
			if err != nil {
				return err
			}

			for k, v := range replacement {
				data = bytes.ReplaceAll(data, []byte(k), []byte(v))
			}

			_, err = file.Write(data)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func paramsProcess(template, appName, modName string) (string, string, string, error) {
	if template != "web" && template != "web+mysql" {
		return "", "", "", fmt.Errorf("only support web、web+mysql, more will be supported in the future")
	}

	if appName == "" {
		appName = "demo"
	}

	if modName == "" {
		modName = appName
	}
	return template, appName, modName, nil
}

func action(c *cli.Context) error {
	tplName, app, mod, err := paramsProcess(template, c.Args().Get(0), modName)
	if err != nil {
		return err
	}
	return copyAndReplace(f, path.Join(".", tplName), app, map[string]string{"${{appName}}": app, "demo-structure": mod})
}

var Command = &cli.Command{Name: "create", Usage: "[-t ${template} [-m ${modName}]] ${appName}", Description: "create a gone app", Flags: flags, Action: action}
