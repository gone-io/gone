package logrus

import (
	"fmt"
	"github.com/gone-io/gone"
)

func NewLogger() (gone.Goner, gone.GonerId) {
	return &defaultLogger{}, gone.IdGoneLogger
}

type defaultLogger struct {
	gone.GonerFlag
}

func (*defaultLogger) Tracef(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (*defaultLogger) Errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (*defaultLogger) Warnf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
func (*defaultLogger) Infof(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}
