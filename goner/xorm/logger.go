package xorm

import (
	"github.com/gone-io/gone/goner/logrus"
	"xorm.io/xorm/log"
)

type dbLogger struct {
	logrus.Logger
	showSql bool
	level   log.LogLevel
}

func (l *dbLogger) Level() log.LogLevel {
	return l.level
}
func (l *dbLogger) SetLevel(level log.LogLevel) {
	l.level = level
}

func (l *dbLogger) ShowSQL(show ...bool) {
	if len(show) > 0 {
		l.showSql = show[0]
	} else {
		l.showSql = true
	}
}
func (l *dbLogger) IsShowSQL() bool {
	return l.showSql
}
