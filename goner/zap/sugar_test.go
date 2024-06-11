package gone_zap

import (
	"github.com/gone-io/gone"
	"testing"
)

func TestNewSugar(t *testing.T) {
	gone.Prepare(Priest).Test(func(log gone.Logger, tracer gone.Tracer) {
		tracer.SetTraceId("", func() {
			log.Info("info log")
			log.Warn("warn log")
			log.Error("error log")
		})
	})
}
