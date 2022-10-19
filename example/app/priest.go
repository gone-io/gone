package app

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/example/app/worker"
	"github.com/gone-io/gone/goner"
)

func Priest(cemetery gone.Cemetery) error {
	_ = goner.BasePriest(cemetery)
	cemetery.Bury(worker.NewPrintWorker())
	cemetery.Bury(worker.NewTimerWorker())
	return nil
}
