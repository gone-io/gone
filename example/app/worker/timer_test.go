package worker_test

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/example/app"
	"github.com/gone-io/gone/example/app/worker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_timerWorker(t *testing.T) {
	gone.
		TestKit(&worker.TimerWorker{}, app.Priest).
		RunAtId(worker.IdTimerWorker, func(worker *worker.TimerWorker) {
			assert.Equal(t, worker.Ttl, 100)
		})
}
