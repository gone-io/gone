package cmux

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/logrus"
)

func NewServer() (gone.Goner, gone.GonerId) {
	return &server{}, gone.IdGoneCumx
}

func Priest(cemetery gone.Cemetery) error {
	_ = logrus.Priest(cemetery)
	_ = config.Priest(cemetery)
	if nil == cemetery.GetTomById(gone.IdGoneCumx) {
		cemetery.Bury(NewServer())
	}
	return nil
}
