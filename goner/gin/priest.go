package gin

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/cmux"
)

func Priest(cemetery gone.Cemetery) error {
	_ = cmux.Priest(cemetery)
	if nil == cemetery.GetTomById(gone.IdGoneGinProxy) {
		cemetery.Bury(NewGinProxy())
	}

	if nil == cemetery.GetTomById(gone.IdGoneGin) {
		cemetery.Bury(NewGinServer())
	}

	if nil == cemetery.GetTomById(gone.IdGoneGinRouter) {
		cemetery.Bury(NewGinRouter())
	}

	if nil == cemetery.GetTomById(gone.IdGoneGinProcessor) {
		cemetery.Bury(NewGinProcessor())
	}

	if nil == cemetery.GetTomById(gone.IdGoneGinResponser) {
		cemetery.Bury(NewGinResponser())
	}
	return nil
}
