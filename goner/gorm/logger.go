package gorm

import (
	"context"
	"errors"
	"fmt"
	"github.com/gone-io/gone"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

type iLogger struct {
	gone.Flag
	log      gone.Logger `gone:"*"`
	LogLevel logger.LogLevel

	SlowThreshold time.Duration `gone:"config,gorm.logger.slow-threshold=200ms"`
}

func (l *iLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *iLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.log.Infof(msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *iLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.log.Warnf(msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *iLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.log.Errorf(msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *iLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound)):
		sql, rows := fc()
		if rows == -1 {
			l.log.Debugf(utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.log.Debugf(utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.log.Debugf(utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.log.Debugf(utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.log.Debugf(utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.log.Debugf(utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
