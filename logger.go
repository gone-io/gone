package gone

import (
	"log"
	"os"
)

type LoggerLevel int8

const (
	DebugLevel LoggerLevel = -1
	InfoLevel  LoggerLevel = 0
	WarnLevel  LoggerLevel = 1
	ErrorLevel LoggerLevel = 2
)

type Logger interface {
	Infof(msg string, args ...any)
	Errorf(msg string, args ...any)
	Warnf(msg string, args ...any)
	Debugf(msg string, args ...any)

	GetLevel() LoggerLevel
	SetLevel(level LoggerLevel)
}

const LoggerName = "gone-logger"

var defaultLog = &defaultLogger{
	level: InfoLevel,
}
var defaultLogInit = false

func GetDefaultLogger() Logger {
	defaultLog.Init()
	return defaultLog
}

type defaultLogger struct {
	Flag
	level    LoggerLevel
	levelStr string `gone:"config,log.level=info"`
}

func (l *defaultLogger) GonerName() string {
	return LoggerName
}

func (l *defaultLogger) Init() {
	if defaultLogInit {
		return
	}
	defaultLogInit = true
	if l.levelStr == "" {
		l.levelStr = os.Getenv("LOG_LEVEL")
	}

	switch l.levelStr {
	case "debug":
		l.level = DebugLevel
	case "warn":
		l.level = WarnLevel
	case "error":
		l.level = ErrorLevel
	default:
		l.level = InfoLevel
	}
}

func (l *defaultLogger) Infof(msg string, args ...any) {
	if l.level <= InfoLevel {
		log.Printf(msg, args...)
	}
}

func (l *defaultLogger) Errorf(msg string, args ...any) {
	if l.level <= ErrorLevel {
		log.Printf(msg, args...)
	}
}

func (l *defaultLogger) Warnf(msg string, args ...any) {
	if l.level <= WarnLevel {
		log.Printf(msg, args...)
	}
}

func (l *defaultLogger) Debugf(msg string, args ...any) {
	if l.level <= DebugLevel {
		log.Printf(msg, args...)
	}
}

func (l *defaultLogger) GetLevel() LoggerLevel {
	return l.level
}

func (l *defaultLogger) SetLevel(level LoggerLevel) {
	l.level = level
}
