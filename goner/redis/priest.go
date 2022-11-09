package redis

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
)

func Priest(cemetery gone.Cemetery) error {
	_ = config.Priest(cemetery)
	_ = logrus.Priest(cemetery)
	_ = tracer.Priest(cemetery)

	if nil == cemetery.GetTomById(gone.IdGoneRedisPool) {
		cemetery.Bury(NewRedisPool())
	}

	if nil == cemetery.GetTomById(IdGoneRedisInner) {
		cemetery.Bury(NewInner())
	}

	if nil == cemetery.GetTomById(gone.IdGoneRedisCache) {
		redisCache, id := NewRedisCache()
		cemetery.Bury(redisCache, id)
		cemetery.Bury(redisCache, gone.IdGoneRedisKey)
	}

	if nil == cemetery.GetTomById(gone.IdGoneRedisLocker) {
		cemetery.Bury(NewRedisLocker())
	}

	if nil == cemetery.GetTomById(gone.IdGoneRedisProvider) {
		cemetery.Bury(NewCacheProvider())
	}
	return nil
}
