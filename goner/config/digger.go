package config

import "github.com/gone-io/gone"

func Digger(cemetery gone.Cemetery) error {
	cemetery.Bury(NewConfig())
	cemetery.Bury(NewConfigure())
	return nil
}
