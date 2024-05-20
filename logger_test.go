package gone

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_defaultLogger_Tracef(t *testing.T) {
	logger, id := NewSimpleLogger()
	assert.Equal(t, IdGoneLogger, string(id))
	l := logger.(*defaultLogger)

	l.Tracef("trace")
	l.Errorf("error")
	l.Warnf("warn")
}
