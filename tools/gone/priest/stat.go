package priest

import (
	log "github.com/sirupsen/logrus"
	"time"
)

var gShowstat bool

func TimeStat(processName string) func() {
	if gShowstat {
		beginTime := time.Now()
		return func() {
			log.Infof("stat <%s> process use time:%v\n", processName, time.Now().Sub(beginTime))
		}
	}
	return func() {}
}
