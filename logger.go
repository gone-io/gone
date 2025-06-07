package gone

import (
	"log"
)

type LoggerLevel int8

const (
	DebugLevel LoggerLevel = -1
	InfoLevel  LoggerLevel = 0
	WarnLevel  LoggerLevel = 1
	ErrorLevel LoggerLevel = 2
)

// Logger Interface which can be injected and provided by gone framework core
type Logger interface {
	Infof(msg string, args ...any)
	Errorf(msg string, args ...any)
	Warnf(msg string, args ...any)
	Debugf(msg string, args ...any)

	GetLevel() LoggerLevel
	SetLevel(level LoggerLevel)
}

const LoggerName = "gone-logger"

func GetDefaultLogger() Logger {
	return &defaultLogger{
		level: new(string),
	}
}

type defaultLogger struct {
	Flag
	level *string `gone:"config,log.level=info"`
}

func (l *defaultLogger) GonerName() string {
	return LoggerName
}

func (l *defaultLogger) Level() LoggerLevel {
	if l.level == nil {
		return InfoLevel
	}

	switch *l.level {
	case "debug":
		return DebugLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

func (l *defaultLogger) Infof(msg string, args ...any) {
	if l.Level() <= InfoLevel {
		log.Printf(msg, args...)
	}
}

func (l *defaultLogger) Errorf(msg string, args ...any) {
	if l.Level() <= ErrorLevel {
		log.Printf(msg, args...)
	}
}

func (l *defaultLogger) Warnf(msg string, args ...any) {
	if l.Level() <= WarnLevel {
		log.Printf(msg, args...)
	}
}

func (l *defaultLogger) Debugf(msg string, args ...any) {
	if l.Level() <= DebugLevel {
		log.Printf(msg, args...)
	}
}

func (l *defaultLogger) GetLevel() LoggerLevel {
	return l.Level()
}

func (l *defaultLogger) SetLevel(level LoggerLevel) {
	switch level {
	case DebugLevel:
		*l.level = "debug"
	case InfoLevel:
		*l.level = "info"
	case WarnLevel:
		*l.level = "warn"
	case ErrorLevel:
		*l.level = "error"
	default:
		*l.level = "info"
	}
}
