package xorm

import (
	"github.com/gone-io/gone"
)

func Priest(cemetery gone.Cemetery) error {
	xormEngine, id, option, g := NewXormEngine()
	cemetery.BuryOnce(xormEngine, id, option, g)
	cemetery.BuryOnce(NewProvider(xormEngine.(*wrappedEngine)))
	return nil
}
