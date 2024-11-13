package schedule

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/redis"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"sync"
	"testing"
	"time"
)

func Test_schedule_AfterRevive(t *testing.T) {
	s := schedule{}
	err := s.AfterRevive()
	assert.Nil(t, err)
	assert.NotNil(t, s.cronTab)
}

type locker struct {
	gone.Flag
	redis.Locker
}

func (l *locker) LockAndDo(key string, fn func(), lockTime, checkPeriod time.Duration) (err error) {
	fn()
	return nil
}

func Test_schedule_Start(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	var mu sync.Mutex
	i := 0
	scheduler := NewMockScheduler(controller)
	scheduler.EXPECT().Cron(gomock.Any()).Do(func(run RunFuncOnceAt) {
		run("0/1 * * * * *", "test", func() {
			println("test")
			mu.Lock()
			i++
			mu.Unlock()
		})
	})

	gone.Prepare(tracer.Priest, logrus.Priest, config.Priest, func(cemetery gone.Cemetery) error {
		cemetery.Bury(&locker{}, gone.IdGoneRedisLocker)
		cemetery.Bury(NewSchedule())
		cemetery.Bury(scheduler)
		return nil
	}).AfterStart(func(in struct {
		s schedule `gone:"*"`
	}) {
		time.Sleep(2 * time.Second)
	}).Run()

	mu.Lock()
	assert.Equal(t, 2, i)
	mu.Unlock()
}
