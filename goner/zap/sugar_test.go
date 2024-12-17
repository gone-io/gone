package gone_zap

import (
	"github.com/gone-io/gone"
	gone_viper "github.com/gone-io/gone/goner/viper"
	"testing"
)

func TestNewSugar(t *testing.T) {
	gone.
		Prepare(
			Priest,
			func(loader gone.Loader) error {
				return gone_viper.Load(loader)
			},
		).
		Test(func(log gone.Logger, tracer gone.Tracer, in struct {
			level string `gone:"config,log.level"`
		}) {
			log.Infof("level:%s", in.level)
			tracer.SetTraceId("", func() {
				log.Debugf("debug log")
				log.Infof("info log")
				log.Warnf("warn log")
				log.Errorf("error log")
			})
		})
}
