package config

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
)

func NewConfig() (gone.Goner, gone.GonerId) {
	return &config{}, gone.IdConfig
}

func NewConfigure() (gone.Goner, gone.GonerId) {
	return &propertiesConfigure{}, gone.IdGoneConfigure
}

func Priest(cemetery gone.Cemetery) error {
	_ = logrus.Priest(cemetery)
	if cemetery.GetTomById(gone.IdConfig) == nil {
		cemetery.Bury(NewConfig())
	}
	if nil == cemetery.GetTomById(gone.IdGoneConfigure) {
		cemetery.Bury(NewConfigure())
	}
	return nil
}
