package config

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/properties"
)

func Priest(cemetery gone.Cemetery) error {
	_ = properties.Priest(cemetery)
	cemetery.BuryOnce(NewConfig())
	return nil
}
