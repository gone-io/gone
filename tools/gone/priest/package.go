package priest

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"path"
	"path/filepath"
)

//for generate Priest function

var flags = []cli.Flag{
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
	&cli.BoolFlag{
		Name:  "w",
		Value: false,
		Usage: "watch files change",
	},
}

func action(c *cli.Context) error {
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
		packageName = path.Base(filepath.Dir(outputFile))
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
}

var Command = &cli.Command{
	Name:        "priest",
	Usage:       "-s ${scanPackageDir} -p ${pkgName} -f ${funcName} -o ${outputFilePath} [-w]",
	Description: "generate gone priest function",
	Flags:       flags,
	Action:      action,
}
