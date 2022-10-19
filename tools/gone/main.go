package main

import (
	"errors"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path"
)

func main() {
	watchFlag := cli.BoolFlag{
		Name:  "w",
		Value: false,
		Usage: "watch files change",
	}

	app := &cli.App{
		Name:    "gone",
		Usage:   "-s ${scan_package_dir} -p ${pkgName} -f ${funcName} -o ${output_dir} [-w]",
		Version: "v0.0.1",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "s",
				Usage:    "scan package dir",
				Required: true,
			},

			&cli.StringFlag{
				Name:     "p",
				Value:    "",
				Usage:    "package name",
				Required: true,
			},

			&cli.StringFlag{
				Name:     "f",
				Value:    "",
				Usage:    "function name",
				Required: true,
			},

			&cli.StringFlag{
				Name:     "o",
				Value:    "",
				Usage:    "output filepath",
				Required: true,
			},

			&cli.BoolFlag{
				Name:  "stat",
				Value: false,
				Usage: "stat process time",
			},
			&watchFlag,
		},
		Action: func(c *cli.Context) error {
			dirs := c.StringSlice("s")
			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			for i := range dirs {
				dirs[i] = path.Join(wd, dirs[i])
			}

			packageName := c.String("p")
			functionName := c.String("f")
			outputFile := c.String("o")

			showStat = c.Bool("stat")

			if outputFile == "" {
				return errors.New("output dir (-o) cannot be empty")
			}

			if packageName == "" {
				packageName = path.Base(path.Dir(outputFile))
			}

			if functionName == "" {
				functionName = "injectLoader"
			}

			loader := autoload{
				scanDir:      dirs,
				packageName:  packageName,
				functionName: functionName,
				outputFile:   outputFile,
			}
			err = loader.fillModuleInfo()
			if err != nil {
				log.Fatalf("loader.fillModuleInfo() err:%v", err)
				return err
			}
			err = loader.firstGenerate()
			if err != nil {
				log.Fatalf("loader.firstGenerate() err:%v", err)
				return err
			}
			if c.Bool("w") {
				log.Println("watch mode...")
				doWatch(loader.reGenerate, dirs)
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("err:", err)
	}
}
