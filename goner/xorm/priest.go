package xorm

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
)

func Priest(cemetery gone.Cemetery) error {
	_ = logrus.Priest(cemetery)
	if nil == cemetery.GetTomById(gone.IdGoneXorm) {
		cemetery.Bury(NewXormEngine())
	}
	return nil
}
