package gone_zap

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
)

func Priest(cemetery gone.Cemetery) error {
	_ = config.Priest(cemetery)

	cemetery.BuryOnce(NewZapLogger())

	theLogger, id, _ := NewSugar()
	return cemetery.ReplaceBury(theLogger, id)
}
