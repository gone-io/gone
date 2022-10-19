package logrus

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/sirupsen/logrus"
)

func NewLogger() (gone.Goner, gone.GonerId) {
	return &logger{
		Logger: logrus.StandardLogger(),
	}, gone.IdGoneLogger
}

func Priest(cemetery gone.Cemetery) error {
	_ = tracer.Priest(cemetery)
	if nil == cemetery.GetTomById(gone.IdGoneLogger) {
		cemetery.Bury(NewLogger())
	}
	return nil
}
