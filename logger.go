package gone

import "fmt"

type SimpleLogger interface {
	Tracef(format string, args ...any)
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
	Infof(format string, args ...any)
}

func NewSimpleLogger() (Goner, GonerId) {
	return &defaultLogger{}, IdGoneLogger
}

type defaultLogger struct {
	Flag
}

func (*defaultLogger) Tracef(format string, args ...any) {
	format = format + "\n"
	fmt.Printf(format, args...)
}

func (*defaultLogger) Errorf(format string, args ...any) {
	format = format + "\n"
	fmt.Printf(format, args...)
}

func (*defaultLogger) Warnf(format string, args ...any) {
	format = format + "\n"
	fmt.Printf(format, args...)
}
func (*defaultLogger) Infof(format string, args ...any) {
	format = format + "\n"
	fmt.Printf(format, args...)
}
