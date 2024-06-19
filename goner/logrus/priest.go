package logrus

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/tracer"
)

func Priest(cemetery gone.Cemetery) error {
	t := cemetery.GetTomById(gone.IdGoneLogger)
	if t != nil && t.GetGoner().(gone.Logger) != gone.GetSimpleLogger() {
		_, ok := t.GetGoner().(*logger)
		if !ok {
			t.GetGoner().(gone.Logger).Warn("logger is loaded, logrus logger not used")
		}
		return nil
	}
	_ = config.Priest(cemetery)
	_ = tracer.Priest(cemetery)
	return cemetery.ReplaceBury(NewLogger())
}
