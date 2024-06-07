package gone_zap

import (
	"github.com/gone-io/gone"
	"testing"
)

func TestNewSugar(t *testing.T) {
	gone.Prepare(Priest).Test(func(log gone.Logger) {
		log.Info("test")
	})
}
