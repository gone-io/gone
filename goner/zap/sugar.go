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
}

func (l *sugar) AfterRevive() error {
	l.SugaredLogger = l.logger.sugar()
	return nil
}

func (l *sugar) Tracef(format string, args ...any) {
	l.Debugf(format, args...)
}

func (l *sugar) Trace(args ...any) {
	l.Debug(args...)
}
func (l *sugar) Traceln(args ...any) {
	l.Debugln(args...)
}

func (l *sugar) Printf(format string, args ...any) {
	l.Infof(format, args...)
}
func (l *sugar) Print(args ...any) {
	l.Info(args...)
}
func (l *sugar) Println(args ...any) {
	l.Infoln(args...)
}

func (l *sugar) Warningf(format string, args ...any) {
	l.Warnf(format, args...)
}
func (l *sugar) Warning(args ...any) {
	l.Warn(args...)
}
func (l *sugar) Warningln(args ...any) {
	l.Warnln(args...)
}
