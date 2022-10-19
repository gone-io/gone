package app

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/example/app/internal"
	"github.com/gone-io/gone/goner"
)

func Priest(cemetery gone.Cemetery) error {
	_ = goner.BasePriest(cemetery)
	_ = internal.Priest(cemetery)
	return nil
}
