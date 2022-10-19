package server

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/example/http-server/internal"
	"github.com/gone-io/gone/goner"
)

func Priest(cemetery gone.Cemetery) error {
	_ = goner.GinPriest(cemetery)
	_ = internal.Priest(cemetery)
	return nil
}
