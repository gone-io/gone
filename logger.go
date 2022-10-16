package gone

import "fmt"

type Logger interface {
	Tracef(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Infof(format string, args ...interface{})
}

type defaultLogger struct {
}

func (*defaultLogger) Tracef(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
func (*defaultLogger) Warnf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
func (*defaultLogger) Infof(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
