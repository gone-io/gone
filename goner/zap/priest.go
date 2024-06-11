package gone_zap

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/tracer"
)

func Priest(cemetery gone.Cemetery) error {
	_ = config.Priest(cemetery)
	_ = tracer.Priest(cemetery)

	cemetery.BuryOnce(NewZapLogger())

	theLogger, id, _ := NewSugar()
	return cemetery.ReplaceBury(theLogger, id)
}
