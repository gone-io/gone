package logrus

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
)

func Priest(cemetery gone.Cemetery) error {
	_ = tracer.Priest(cemetery)
	if nil == cemetery.GetTomById(gone.IdGoneLogger) {
		logger, id := NewLogger()
		cemetery.Bury(logger, id)

		tombs := cemetery.GetTomByType(gone.GetInterfaceType(new(gone.DefaultLoggerUser)))
		for _, tomb := range tombs {
			goner := tomb.GetGoner()
			user := goner.(gone.DefaultLoggerUser)
			_ = user.SetLogger(logger.(gone.Logger))
		}
	}
	return nil
}
