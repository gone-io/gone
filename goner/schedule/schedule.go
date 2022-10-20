package schedule

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/redis"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/robfig/cron/v3"
	"time"
)

func NewSchedule() (gone.Angel, gone.GonerId) {
	return &schedule{}, gone.IdGoneSchedule
}

type schedule struct {
	gone.Flag
	cronTab       *cron.Cron
	logrus.Logger `gone:"gone-logger"`
	tracer        tracer.Tracer `gone:"gone-tracer"`
	Locker        redis.Locker  `gone:"gone-redis-locker"`
	schedulers    []Scheduler   `gone:"*"`

	lockTime    time.Duration `gone:"config,schedule.lockTime,default=10s"`
	checkPeriod time.Duration `gone:"config,schedule.checkPeriod,default=2s"`
}

func (s *schedule) Start(gone.Cemetery) error {
	s.cronTab = cron.New(cron.WithSeconds())

	for _, o := range s.schedulers {
		o.Cron(func(spec string, jobName JobName, fn func()) {
			lockKey := fmt.Sprintf("lock-job:%s", jobName)

			_, err := s.cronTab.AddFunc(spec, func() {
				s.tracer.RecoverSetTraceId("", func() {
					err := s.Locker.LockAndDo(lockKey, fn, s.lockTime, s.checkPeriod)
					if err != nil {
						s.Warnf("cron get lock err:%v", err)
					}
				})
			})

			if err != nil {
				panic("cron.AddFunc for " + string(jobName) + " err:" + err.Error())
			}
			s.Infof("Add cron item: %s :%s", spec, jobName)
		})
	}
	s.cronTab.Start()
	return nil
}

func (s *schedule) Stop(gone.Cemetery) error {
	s.cronTab.Stop()
	return nil
}
