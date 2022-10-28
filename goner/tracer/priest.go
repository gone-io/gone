package tracer

import (
	"github.com/gone-io/gone"
)

func Priest(cemetery gone.Cemetery) error {
	if nil == cemetery.GetTomById(gone.IdGoneTracer) {
		cemetery.Bury(NewTracer())
	}
	return nil
}
