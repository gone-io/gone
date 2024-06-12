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
			log.Trace("trace log")
			log.Tracef("trace log: %d", 1)
			log.Traceln("trace log")

			log.Printf("%s", "print log")
			log.Print("print log")
			log.Println("print log")

			log.Warningf("warning log: %d", 1)
			log.Warning("warning log")
			log.Warningln("warning log")
		})
	})
}
