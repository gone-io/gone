package priest

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

func CreateCommand() *cli.Command {
	return &cli.Command{
		Name:        "priest",
		Usage:       "-s ${scanPackageDir} -p ${pkgName} -f ${funcName} -o ${outputFilePath} [-w]",
		Description: "generate gone priest function",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "scan-dir",
				Aliases:  []string{"s"},
				Usage:    "scan package dir",
				Required: true,
			},

			&cli.StringFlag{
				Name:     "package",
				Aliases:  []string{"p"},
				Value:    "",
				Usage:    "package name of generated code",
				Required: true,
			},

			&cli.StringFlag{
				Name:     "function",
				Aliases:  []string{"f"},
				Value:    "",
				Usage:    "function name of generated code",
				Required: true,
			},

			&cli.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Value:    "",
				Usage:    "output filepath of generated code",
				Required: true,
			},

			&cli.BoolFlag{
				Name:  "stat",
				Value: false,
				Usage: "is stat process time",
			},
			&cli.BoolFlag{
				Name:    "watch",
				Aliases: []string{"w"},
				Value:   false,
				Usage:   "watch files change, and generate code when any files changed",
			},
		},
		Action: action,
	}
}

func action(c *cli.Context) error {
	return doAction(
		c.StringSlice("scan-dir"),
		c.String("package"),
		c.String("function"),
		c.String("output"),
		c.Bool("stat"),
		c.Bool("watch"),
	)
}

func doAction(
	dirs []string,
	packageName, functionName, outputFile string,
	showStat, isWatch bool,
) error {
	gShowstat = showStat

	if len(dirs) == 0 {
		wd, _ := os.Getwd()
		dirs = append(dirs, wd)
	}

	for i := range dirs {
		dirs[i], _ = filepath.Abs(dirs[i])
	}

	if !filepath.IsAbs(outputFile) {
		outputFile, _ = filepath.Abs(outputFile)
	}

	loader := autoload{
		scanDir:      dirs,
		packageName:  packageName,
		functionName: functionName,
		outputFile:   outputFile,
	}
	err := loader.fillModuleInfo()
	if err != nil {
		log.Fatalf("loader.fillModuleInfo() err:%v", err)
		return err
	}
	err = loader.firstGenerate()
	if err != nil {
		log.Fatalf("loader.firstGenerate() err:%v", err)
		return err
	}

	if isWatch {
		log.Println("watch mode...")
		doWatch(loader.reGenerate, dirs, outputFile)
	}
	return nil
}
