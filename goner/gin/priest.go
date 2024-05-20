package gin

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/cmux"
)

func ginPriest(cemetery gone.Cemetery) error {
	arr := []func() (gone.Goner, gone.GonerId){
		NewGinProxy,
		NewGinRouter,
		NewGinProcessor,
		NewGinResponser,
		NewGinServer,
		NewHttInjector,
	}

	for _, f := range arr {
		goner, id := f()
		gone.CheckAndBury(cemetery, goner, id)
	}
	return nil
}

func Priest(cemetery gone.Cemetery) error {
	_ = cmux.Priest(cemetery)
	_ = ginPriest(cemetery)
	return nil
}
