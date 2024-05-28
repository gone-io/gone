package logrus

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
)

func Priest(cemetery gone.Cemetery) error {
	_ = tracer.Priest(cemetery)
	if nil == cemetery.GetTomById(gone.IdGoneLogger) {
		theLogger, id, option := NewLogger()
		cemetery.Bury(theLogger, id, option)

		tombs := cemetery.GetTomByType(gone.GetInterfaceType(new(gone.DefaultLogger)))
		for _, tomb := range tombs {
			goner := tomb.GetGoner()
			log := goner.(gone.DefaultLogger)
			_ = log.SetLogger(theLogger.(gone.SimpleLogger))
		}
	}
	return nil
}
