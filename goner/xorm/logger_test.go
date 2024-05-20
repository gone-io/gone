package xorm

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"xorm.io/xorm/log"
)

func Test_dbLogger_Level(t *testing.T) {
	logger := dbLogger{}

	logger.SetLevel(log.LOG_INFO)
	level := logger.Level()
	assert.Equal(t, log.LOG_INFO, level)

	logger.ShowSQL(false)
	assert.False(t, logger.IsShowSQL())
	logger.ShowSQL()
	assert.True(t, logger.IsShowSQL())
}
