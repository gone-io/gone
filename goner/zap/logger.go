package gone_zap

import (
	"fmt"
	"github.com/gone-io/gone"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"reflect"
	"strings"
)
import "gopkg.in/natefinch/lumberjack.v2"

func NewZapLogger() (gone.Goner, gone.GonerId, gone.IsDefault) {
	return &log{}, "zap", true
}

type log struct {
	gone.Flag
	*zap.Logger
	tracer gone.Tracer `gone:"*"`

	level             string `gone:"config,log.level,default=info"`
	enableTraceId     bool   `gone:"config,log.enable-trace-id,default=true"`
	disableStacktrace bool   `gone:"config,log.disable-stacktrace,default=false"`
	stackTraceLevel   string `gone:"config,log.stacktrace-level,default=error"`

	reportCaller bool   `gone:"config,log.report-caller,default=true"`
	encoder      string `gone:"config,log.encoder,default=console"`

	output    string `gone:"config,log.output,default=stdout"`
	errOutput string `gone:"config,log.error-output,default=stderr"`

	rotationOutput      string `gone:"config,log.rotation.output"`
	rotationErrorOutput string `gone:"config,log.rotation.error-output"`
	rotationMaxSize     int    `gone:"config,log.rotation.max-size,default=100"`
	rotationMaxFiles    int    `gone:"config,log.rotation.max-files,default=10"`
	rotationMaxAge      int    `gone:"config,log.rotation.max-age,default=30"`
	rotationLocalTime   bool   `gone:"config,log.rotation.local-time,default=true"`
	rotationCompress    bool   `gone:"config,log.rotation.compress,default=false"`
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
		l.Logger, err = l.Build()
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *log) Build() (*zap.Logger, error) {
	outputs := strings.Split(l.output, ",")
	sink, closeOut, err := zap.Open(outputs...)
	if err != nil {
		return nil, gone.ToError(err)
	}

	if l.rotationOutput != "" {
		rotationWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   l.rotationOutput,
			MaxSize:    l.rotationMaxSize, // megabytes
			MaxBackups: l.rotationMaxFiles,
			MaxAge:     l.rotationMaxAge, // days
			Compress:   l.rotationCompress,
		})

		sink = zap.CombineWriteSyncers(sink, rotationWriter)
	}

	errOutputs := strings.Split(l.errOutput, ",")
	var errSink zapcore.WriteSyncer
	if len(errOutputs) > 0 {
		errSink, _, err = zap.Open(errOutputs...)
		if err != nil {
			closeOut()
			return nil, gone.ToError(err)
		}
	}

	if l.rotationErrorOutput != "" {
		rotationWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   l.rotationErrorOutput,
			MaxSize:    l.rotationMaxSize, // megabytes
			MaxBackups: l.rotationMaxFiles,
			MaxAge:     l.rotationMaxAge, // days
			Compress:   l.rotationCompress,
		})
		if errSink == nil {
			errSink = rotationWriter
		} else {
			errSink = zap.CombineWriteSyncers(errSink, rotationWriter)
		}
	}
	var encoder zapcore.Encoder
	if l.encoder == "console" {
		config := zap.NewDevelopmentEncoderConfig()
		config.ConsoleSeparator = "|"
		encoder = zapcore.NewConsoleEncoder(config)
	} else {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	}

	if l.enableTraceId {
		encoder = NewTraceEncoder(encoder, l.tracer)
	}

	core := zapcore.NewCore(
		encoder,
		sink,
		parseLevel(l.level),
	)

	var opts []Option

	if errSink != nil {
		opts = append(opts, zap.ErrorOutput(errSink))
	}
	if !l.disableStacktrace {
		opts = append(opts, zap.AddStacktrace(parseLevel(l.stackTraceLevel)))
	}

	if l.reportCaller {
		opts = append(opts, zap.AddCaller())
	}

	logger := zap.New(core)
	if len(opts) > 0 {
		logger = logger.WithOptions(opts...)
	}
	return logger, nil
}

func parseLevel(level string) zapcore.Level {
	switch level {
	default:
		return zap.InfoLevel
	case "debug", "trace":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	}
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
