package config

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
)

func NewConfig() (gone.Vampire, gone.GonerId, gone.GonerOption) {
	return &config{}, gone.IdConfig, gone.IsDefault(true)
}

func NewConfigure() (gone.Goner, gone.GonerId, gone.GonerOption) {
	return &propertiesConfigure{}, gone.IdGoneConfigure, gone.IsDefault(true)
}

func Priest(cemetery gone.Cemetery) error {
	_ = logrus.Priest(cemetery)

	cemetery.
		BuryOnce(NewConfig()).
		BuryOnce(NewConfigure())
	return nil
}
