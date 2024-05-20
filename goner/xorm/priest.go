package xorm

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/logrus"
)

func Priest(cemetery gone.Cemetery) error {
	_ = config.Priest(cemetery)
	_ = logrus.Priest(cemetery)
	xormEngine, id := NewXormEngine()
	gone.CheckAndBury(cemetery, xormEngine, id)
	return nil
}
