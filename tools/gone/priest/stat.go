package priest

import (
	log "github.com/sirupsen/logrus"
	"time"
)

var showStat = false

func TimeStat(processName string) func() {
	if showStat {
		beginTime := time.Now()
		return func() {
			log.Infof("stat <%s> process use time:%v\n", processName, time.Now().Sub(beginTime))
		}
	}
	return func() {}
}
