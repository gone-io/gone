package xorm

import (
	"github.com/gone-io/gone"
)

func Priest(cemetery gone.Cemetery) error {
	if nil == cemetery.GetTomById(gone.IdGoneXorm) {
		cemetery.Bury(NewXormEngine())
	}
	return nil
}
