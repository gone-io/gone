package gone_viper

import (
	"github.com/gone-io/gone"
)

func Priest(cemetery gone.Cemetery) error {
	return cemetery.ReplaceBury(NewConfigure())
}
