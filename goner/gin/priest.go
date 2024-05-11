package gin

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/cmux"
)

func ginPriest(cemetery gone.Cemetery) error {
	if nil == cemetery.GetTomById(gone.IdGoneGinProxy) {
		cemetery.Bury(NewGinProxy())
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

	if nil == cemetery.GetTomById(gone.IdGoneGin) {
		cemetery.Bury(NewGinServer())
	}
	if nil == cemetery.GetTomById(gone.IdHttpInjector) {
		cemetery.Bury(NewHttInjector())
	}
	return nil
}

func Priest(cemetery gone.Cemetery) error {
	_ = cmux.Priest(cemetery)
	_ = ginPriest(cemetery)
	return nil
}
