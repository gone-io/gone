package gone_zap

import (
	"github.com/gone-io/gone"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field = zapcore.Field
type Option = zap.Option
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Panic(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	Named(s string) Logger
	WithOptions(opts ...Option) Logger
	With(fields ...Field) Logger
	Sugar() gone.Logger

	sugar() *zap.SugaredLogger
}
