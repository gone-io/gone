package app

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/gone-io/gone/templates"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"strings"
)

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
