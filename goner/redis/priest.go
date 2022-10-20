package redis

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
)

func Priest(cemetery gone.Cemetery) error {
	_ = logrus.Priest(cemetery)
	if nil == cemetery.GetTomById(gone.IdGoneRedisPool) {
		cemetery.Bury(NewRedisPool())
	}

	if nil == cemetery.GetTomById(IdGoneRedisInner) {
		cemetery.Bury(NewInner())
	}

	if nil == cemetery.GetTomById(gone.IdGoneRedisCache) {
		cemetery.Bury(NewRedisCache())
	}

	if nil == cemetery.GetTomById(gone.IdGoneRedisLocker) {
		cemetery.Bury(NewRedisLocker())
	}
	return nil
}
