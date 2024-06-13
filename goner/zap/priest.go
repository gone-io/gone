package gone_zap

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/tracer"
)

func Priest(cemetery gone.Cemetery) error {
	t := cemetery.GetTomById(gone.IdGoneLogger)
	if t != nil && t.GetGoner().(gone.Logger) != gone.GetSimpleLogger() {
		t.GetGoner().(gone.Logger).Warn("logger is loaded, zap logger not used")
		return nil
	}

	_ = config.Priest(cemetery)
	_ = tracer.Priest(cemetery)

	cemetery.BuryOnce(NewZapLogger())

	return cemetery.ReplaceBury(NewSugar())
}
