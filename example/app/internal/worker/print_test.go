package worker_test

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/example/app"
	"github.com/gone-io/gone/example/app/internal/worker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_printWorker(t *testing.T) {
	t.Run("content", func(t *testing.T) {
		gone.
			Test(func(printWorker worker.PrintWorker) {
				assert.Equal(t, printWorker.GetContent(), "ok")
			}, app.Priest)
	})
}
