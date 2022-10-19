package tracer

import (
	"github.com/gone-io/gone"
	"github.com/jtolds/gls"
)

func NewTracer() (gone.Goner, gone.GonerId) {
	return &tracer{
		mgr:        gls.NewContextManager(),
		traceIdKey: gls.GenSym(),
	}, gone.IdGoneTracer
}

func Priest(cemetery gone.Cemetery) error {
	if nil == cemetery.GetTomById(gone.IdGoneTracer) {
		cemetery.Bury(NewTracer())
	}
	return nil
}
