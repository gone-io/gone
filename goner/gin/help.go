package gin

import (
	"fmt"
	"time"
)

type timeUseRecord struct {
	UseTime time.Duration
	Count   int64
}

var mapRecord = make(map[string]*timeUseRecord)

// TimeStat record the time of function and avg time
func TimeStat(name string, start time.Time, logs ...func(format string, args ...any)) {
	since := time.Since(start)
	if mapRecord[name] == nil {
		mapRecord[name] = &timeUseRecord{}
	}
	mapRecord[name].UseTime += since
	mapRecord[name].Count++

	var log func(format string, args ...any)
	if len(logs) == 0 {
		log = func(format string, args ...any) {
			fmt.Printf(format, args...)
		}
	} else {
		log = logs[0]
	}

	log("%s executed %v times, took %v, avg: %v\n",
		name,
		mapRecord[name].Count,
		mapRecord[name].UseTime,
		mapRecord[name].UseTime/time.Duration(mapRecord[name].Count),
	)
}
