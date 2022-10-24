package worker_test

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/example/app"
	"github.com/gone-io/gone/example/app/internal/worker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_timerWorker(t *testing.T) {
	gone.
		TestAt(worker.IdTimerWorker, func(worker *worker.TimerWorker) {
			assert.Equal(t, worker.Ttl, 100)
		}, app.Priest)
}
