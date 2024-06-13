package properties

import (
	"github.com/gone-io/gone"
)

func Priest(cemetery gone.Cemetery) error {
	cemetery.BuryOnce(NewConfigure())
	return nil
}
