package tracer

import (
	"github.com/gone-io/gone"
)

func Priest(cemetery gone.Cemetery) error {
	cemetery.BuryOnce(NewTracer())
	return nil
}
