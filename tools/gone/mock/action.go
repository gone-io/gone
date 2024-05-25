package mock

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

func Action(ctx *cli.Context) error {
	outfilePath := ctx.String("o")
	err := os.MkdirAll(filepath.Dir(outfilePath), os.ModePerm)
	if err != nil {
		return err
	}

	outFile, err := os.Create(outfilePath)
	if err != nil {
		return err
	}

	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			log.Error(err)
		}
	}(outFile)

	if isInputFromPipe() {
		return patchMock(os.Stdin, outFile)
	} else {
		file, e := getFile(ctx.String("f"))
		if e != nil {
			return e
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Error(err)
			}
		}(file)
		return patchMock(file, outFile)
	}
}
