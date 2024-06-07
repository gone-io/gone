package gone_zap

import (
	"fmt"
	"github.com/gone-io/gone"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"reflect"
)

func NewZapLogger() (gone.Goner, gone.GonerId, gone.IsDefault) {
	return &log{}, "zap", true
}

type log struct {
	gone.Flag
	*zap.Logger

	level        string `gone:"config,log.level,default=info"`
	reportCaller bool   `gone:"config,log.report-caller,default=true"`
	output       string `gone:"config,log.output,default=stdout"`
	format       string `gone:"config,log.format,default=text"`
}

func (l *log) Named(s string) Logger {
	if s == "" {
		return l
	}
	return &log{Logger: l.Logger.Named(s)}
}
func (l *log) WithOptions(opts ...Option) Logger {
	if len(opts) == 0 {
		return l
	}
	return &log{Logger: l.Logger.WithOptions(opts...)}
}
func (l *log) With(fields ...Field) Logger {
	if len(fields) == 0 {
		return l
	}
	return &log{Logger: l.Logger.With(fields...)}
}
func (l *log) Sugar() gone.Logger {
	return &sugar{SugaredLogger: l.Logger.Sugar()}
}

func (l *log) sugar() *zap.SugaredLogger {
	if l.Logger == nil {
		_ = l.AfterRevive()
	}
	return l.Logger.Sugar()
}

func (l *log) AfterRevive() (err error) {
	if l.Logger == nil {
		cfg := zap.Config{
			Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
			Development: false,
			Encoding:    "json",
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:        "time",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "", // 不记录日志调用位置
				FunctionKey:    zapcore.OmitKey,
				MessageKey:     "message",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.RFC3339TimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
			OutputPaths:      []string{"stdout", "testdata/test.log"},
			ErrorOutputPaths: []string{"testdata/error.log"},
		}

		l.Logger, err = cfg.Build()
		//l.Logger, err = zap.NewProduction()

		if err != nil {
			return gone.ToError(err)
		}
	}
	return nil
}
func (l *log) Start(gone.Cemetery) error {
	return nil
}
func (l *log) Stop(gone.Cemetery) error {
	defer l.Logger.Sync()
	return nil
}

func (l *log) Suck(conf string, v reflect.Value, field reflect.StructField) error {
	t := field.Type

	goner := l.Named(conf)

	if gone.IsCompatible(t, goner) {
		v.Set(reflect.ValueOf(goner))
		return nil
	}
	sLogger := goner.Sugar()
	if gone.IsCompatible(t, sLogger) {
		v.Set(reflect.ValueOf(sLogger))
		return nil
	}

	return gone.NewInnerError(
		fmt.Sprintf("the attribute(%s) do not support type(%s.%s) for gone zap tag; use gone.Logger or gone_zap.Logger instead", field.Name, t.PkgPath(), t.Name()),
		gone.NotCompatible,
	)
}
