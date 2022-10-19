package app

import (
	"github.com/gone-io/gone"
	worker2 "github.com/gone-io/gone/example/app/internal/worker"
	"github.com/gone-io/gone/goner"
)

func Priest(cemetery gone.Cemetery) error {
	_ = goner.BasePriest(cemetery)
	cemetery.Bury(worker2.NewPrintWorker())
	cemetery.Bury(worker2.NewTimerWorker())
	return nil
}
