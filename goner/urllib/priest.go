package urllib

import (
	"github.com/gone-io/gone"
)

func Priest(cemetery gone.Cemetery) error {
	cemetery.BuryOnce(NewReq())
	return nil
}
