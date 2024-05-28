package urllib

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
)

func Priest(cemetery gone.Cemetery) error {
	_ = tracer.Priest(cemetery)
	cemetery.BuryOnce(NewReq())
	return nil
}
