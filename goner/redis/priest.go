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

	cemetery.
		BuryOnce(NewRedisPool()).
		BuryOnce(NewInner()).
		BuryOnce(NewRedisCache()).
		BuryOnce(NewRedisKey()).
		BuryOnce(NewRedisLocker()).
		BuryOnce(NewCacheProvider())
	return nil
}
