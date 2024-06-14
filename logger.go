package gone

import (
	"log"
)

var _defaultLogger = &defaultLogger{Logger: new(log.Logger)}

func NewSimpleLogger() (Goner, GonerId, IsDefault) {
	return _defaultLogger, IdGoneLogger, true
}

func GetSimpleLogger() Logger {
	return _defaultLogger
}

type defaultLogger struct {
	Flag
	*log.Logger
}

func (l *defaultLogger) Tracef(format string, args ...any) {
	log.Printf(format, args...)
}
func (l *defaultLogger) Debugf(format string, args ...any) {
	log.Printf(format, args...)
}
func (l *defaultLogger) Infof(format string, args ...any) {
	log.Printf(format, args...)
}
func (l *defaultLogger) Warnf(format string, args ...any) {
	log.Printf(format, args...)
}

func (l *defaultLogger) Errorf(format string, args ...any) {
	log.Printf(format, args...)
}

func (l *defaultLogger) Trace(args ...any) {
	log.Print(args...)
}
func (l *defaultLogger) Debug(args ...any) {
	log.Print(args...)
}
func (l *defaultLogger) Info(args ...any) {
	log.Print(args...)
}
func (l *defaultLogger) Warn(args ...any) {
	log.Print(args...)
}

func (l *defaultLogger) Error(args ...any) {
	log.Print(args...)
}

func (l *defaultLogger) Traceln(args ...any) {
	log.Println(args...)
}
func (l *defaultLogger) Debugln(args ...any) {
	log.Println(args...)
}
func (l *defaultLogger) Infoln(args ...any) {
	log.Println(args...)
}
func (l *defaultLogger) Warnln(args ...any) {
	log.Println(args...)
}

func (l *defaultLogger) Errorln(args ...any) {
	log.Println(args...)
}
