package gone

import "fmt"

type Logger interface {
	Tracef(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Infof(format string, args ...interface{})
}

func NewSimpleLogger() (Goner, GonerId) {
	return &defaultLogger{}, IdGoneLogger
}

type defaultLogger struct {
	Flag
}

func (*defaultLogger) Tracef(format string, args ...interface{}) {
	format = format + "\n"
	fmt.Printf(format, args...)
}

func (*defaultLogger) Errorf(format string, args ...interface{}) {
	format = format + "\n"
	fmt.Printf(format, args...)
}

func (*defaultLogger) Warnf(format string, args ...interface{}) {
	format = format + "\n"
	fmt.Printf(format, args...)
}
func (*defaultLogger) Infof(format string, args ...interface{}) {
	format = format + "\n"
	fmt.Printf(format, args...)
}
