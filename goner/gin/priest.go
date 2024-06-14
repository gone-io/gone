package gin

import (
	"github.com/gone-io/gone"
)

func Priest(cemetery gone.Cemetery) error {
	cemetery.
		BuryOnce(NewGinProxy()).
		BuryOnce(NewGinRouter()).
		BuryOnce(NewGinProcessor()).
		BuryOnce(NewGinResponser()).
		BuryOnce(NewGinServer()).
		BuryOnce(NewHttInjector())
	return nil
}
