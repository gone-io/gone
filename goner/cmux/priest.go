package cmux

import (
	"github.com/gone-io/gone"
)

func Priest(cemetery gone.Cemetery) error {
	cemetery.BuryOnce(NewServer())
	return nil
}
