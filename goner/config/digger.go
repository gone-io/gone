package config

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
)

func Digger(cemetery gone.Cemetery) error {
	logger := cemetery.GetTomById(gone.IdGoneLogger)
	if logger == nil {
		cemetery.Bury(logrus.NewLogger())
	}
	cemetery.Bury(NewConfig())
	cemetery.Bury(NewConfigure())
	return nil
}
