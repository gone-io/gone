package schedule

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/redis"
	"github.com/robfig/cron/v3"
	"time"
)

var load = gone.OnceLoad(func(loader gone.Loader) error {
	return loader.Load(&schedule{})
})

func Load(loader gone.Loader) error {
	return load(loader)
}

// Priest Deprecated, use Load instead
func Priest(loader gone.Loader) error {
	return Load(loader)
}

type schedule struct {
	gone.Flag
	cronTab     *cron.Cron
	gone.Logger `gone:"gone-logger"`
	tracer      gone.Tracer  `gone:"gone-tracer"`
	locker      redis.Locker `gone:"gone-redis-locker"`
	schedulers  []Scheduler  `gone:"*"`

	lockTime    time.Duration `gone:"config,schedule.lockTime,default=10s"`
	checkPeriod time.Duration `gone:"config,schedule.checkPeriod,default=2s"`
}

func (s *schedule) Init() {
	s.cronTab = cron.New(cron.WithSeconds())
}

func (s *schedule) Start() error {
	for _, o := range s.schedulers {
		o.Cron(func(spec string, jobName JobName, fn func()) {
			lockKey := fmt.Sprintf("lock-job:%s", jobName)

			_, err := s.cronTab.AddFunc(spec, func() {
				s.tracer.RecoverSetTraceId("", func() {
					err := s.locker.LockAndDo(lockKey, fn, s.lockTime, s.checkPeriod)
					if err != nil {
						s.Warnf("cron get lock err:%v", err)
					}
				})
			})

			if err != nil {
				panic("cron.AddFunc for " + string(jobName) + " err:" + err.Error())
			}
			s.Infof("Add cron item: %s => %s : %s", spec, jobName, gone.GetFuncName(fn))
		})
	}
	s.cronTab.Start()
	return nil
}

func (s *schedule) Stop() error {
	s.cronTab.Stop()
	return nil
}
