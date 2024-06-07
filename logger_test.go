package gone

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_defaultLogger_Tracef(t *testing.T) {
	Prepare().Test(func(log Logger) {
		logger, _, _ := NewSimpleLogger()
		assert.Equal(t, logger, log)

		log.Tracef("format: %s", "test")
		log.Debugf("format: %s", "test")
		log.Infof("format: %s", "test")
		log.Warnf("format: %s", "test")
		log.Warningf("format: %s", "test")
		log.Errorf("format: %s", "test")

		log.Trace("test")
		log.Debug("test")
		log.Info("test")
		log.Warn("test")
		log.Warning("test")
		log.Error("test")

		log.Traceln("test")
		log.Debugln("test")
		log.Infoln("test")
		log.Warnln("test")
		log.Warningln("test")
		log.Errorln("test")
	})
}
