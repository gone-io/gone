package config

import (
	"github.com/gone-io/gone"
)

func Priest(cemetery gone.Cemetery) error {
	cemetery.BuryOnce(NewConfigure()).BuryOnce(NewConfig())
	return nil
}
