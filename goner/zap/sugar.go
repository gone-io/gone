package gone_zap

import (
	"github.com/gone-io/gone"
	"go.uber.org/zap"
)

func NewSugar() (gone.Goner, gone.GonerId, gone.IsDefault) {
	logger, err := zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		panic(gone.ToError(err))
	}
	return &sugar{
		SugaredLogger: logger.Sugar(),
	}, gone.IdGoneLogger, true
}

type sugar struct {
	gone.Flag
	*zap.SugaredLogger
	logger Logger `gone:"*"`

	inner *zap.SugaredLogger
}

func (l *sugar) AfterRevive() error {
	l.SugaredLogger = l.logger.sugar()
	l.inner = l.WithOptions(zap.AddCallerSkip(1))
	return nil
}

func (l *sugar) Tracef(format string, args ...any) {
	l.inner.Debugf(format, args...)
}

func (l *sugar) Trace(args ...any) {
	l.inner.Debug(args...)
}
func (l *sugar) Traceln(args ...any) {
	l.inner.Debugln(args...)
}

func (l *sugar) Printf(format string, args ...any) {
	l.inner.Infof(format, args...)
}
func (l *sugar) Print(args ...any) {
	l.inner.Info(args...)
}
func (l *sugar) Println(args ...any) {
	l.inner.Infoln(args...)
}

func (l *sugar) Warningf(format string, args ...any) {
	l.inner.Warnf(format, args...)
}
func (l *sugar) Warning(args ...any) {
	l.inner.Warn(args...)
}
func (l *sugar) Warningln(args ...any) {
	l.inner.Warnln(args...)
}
