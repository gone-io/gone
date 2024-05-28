package schedule

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/redis"
	"github.com/gone-io/gone/goner/tracer"
)

func Priest(cemetery gone.Cemetery) error {
	_ = tracer.Priest(cemetery)
	_ = logrus.Priest(cemetery)
	_ = redis.Priest(cemetery)

	cemetery.BuryOnce(NewSchedule())
	return nil
}
