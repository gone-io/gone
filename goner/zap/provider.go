package gone_zap

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"strings"
)

// Load load zap logger
var load = gone.OnceLoad(func(loader gone.Loader) error {
	err := tracer.Load(loader)
	if err != nil {
		return err
	}

	err = loader.Load(&zapLoggerProvider{})
	if err != nil {
		return err
	}
	err = loader.Load(&sugarProvider{})
	if err != nil {
		return err
	}
	return loader.Load(&sugar{}, gone.IsDefault(new(gone.Logger)), gone.ForceReplace())
})

func Load(loader gone.Loader) error {
	return load(loader)
}

// Priest Deprecated, use Load instead
func Priest(loader gone.Loader) error {
	return Load(loader)
}

type wrappedLogger struct {
	*zap.Logger
}

func (l *wrappedLogger) sugar() *zap.SugaredLogger {
	return l.Logger.Sugar()
}

func (l *wrappedLogger) Named(s string) Logger {
	if s == "" {
		return l
	}
	return &wrappedLogger{Logger: l.Logger.Named(s)}
}
func (l *wrappedLogger) WithOptions(opts ...Option) Logger {
	if len(opts) == 0 {
		return l
	}
	return &wrappedLogger{Logger: l.Logger.WithOptions(opts...)}
}
func (l *wrappedLogger) With(fields ...Field) Logger {
	if len(fields) == 0 {
		return l
	}
	return &wrappedLogger{Logger: l.Logger.With(fields...)}
}
func (l *wrappedLogger) Sugar() gone.Logger {
	SugaredLogger := l.Logger.Sugar()
	return &sugar{
		SugaredLogger: SugaredLogger,
	}
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

type zapLoggerProvider struct {
	gone.Flag

	level             string `gone:"config,log.level,default=info"`
	enableTraceId     bool   `gone:"config,log.enable-trace-id,default=true"`
	disableStacktrace bool   `gone:"config,log.disable-stacktrace,default=false"`
	stackTraceLevel   string `gone:"config,log.stacktrace-level,default=error"`

	reportCaller bool   `gone:"config,log.report-caller,default=true"`
	encoder      string `gone:"config,log.encoder,default=console"`

	output    string `gone:"config,log.output,default=stdout"`
	errOutput string `gone:"config,log.error-output"`

	rotationOutput      string `gone:"config,log.rotation.output"`
	rotationErrorOutput string `gone:"config,log.rotation.error-output"`
	rotationMaxSize     int    `gone:"config,log.rotation.max-size,default=100"`
	rotationMaxFiles    int    `gone:"config,log.rotation.max-files,default=10"`
	rotationMaxAge      int    `gone:"config,log.rotation.max-age,default=30"`
	rotationLocalTime   bool   `gone:"config,log.rotation.local-time,default=true"`
	rotationCompress    bool   `gone:"config,log.rotation.compress,default=false"`

	before      gone.BeforeStop `gone:"*"`
	tracer      gone.Tracer     `gone:"*"`
	zapLogger   *zap.Logger
	atomicLevel zap.AtomicLevel
}

func (s *zapLoggerProvider) Provide(tagConf string) (*zap.Logger, error) {
	_, keys := gone.TagStringParse(tagConf)

	if len(keys) > 0 {
		if keys[0] != "" {
			return s.zapLogger.Named(keys[0]), nil
		}
	}
	return s.zapLogger, nil
}

func (s *zapLoggerProvider) Init() error {
	if s.zapLogger == nil {
		logger, err := s.create()
		if err != nil {
			return gone.ToError(err)
		}
		s.zapLogger = logger
		s.before(func() {
			err := s.zapLogger.Sync()
			if err != nil {
				gone.GetDefaultLogger().Errorf("failed to sync logger:%v", err)
			}
		})
	}
	return nil
}
func (s *zapLoggerProvider) SetLevel(level zapcore.Level) {
	s.atomicLevel.SetLevel(level)
}

func (s *zapLoggerProvider) create() (*zap.Logger, error) {
	outputs := strings.Split(s.output, ",")
	sink, closeOut, err := zap.Open(outputs...)
	if err != nil {
		return nil, gone.ToError(err)
	}

	if s.rotationOutput != "" {
		rotationWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   s.rotationOutput,
			MaxSize:    s.rotationMaxSize, // megabytes
			MaxBackups: s.rotationMaxFiles,
			MaxAge:     s.rotationMaxAge, // days
			Compress:   s.rotationCompress,
		})

		sink = zap.CombineWriteSyncers(sink, rotationWriter)
	}

	errOutputs := strings.Split(s.errOutput, ",")
	var errSink zapcore.WriteSyncer
	if s.errOutput != "" && len(errOutputs) > 0 {
		errSink, _, err = zap.Open(errOutputs...)
		if err != nil {
			closeOut()
			return nil, gone.ToError(err)
		}
	}

	if s.rotationErrorOutput != "" {
		rotationWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   s.rotationErrorOutput,
			MaxSize:    s.rotationMaxSize, // megabytes
			MaxBackups: s.rotationMaxFiles,
			MaxAge:     s.rotationMaxAge, // days
			Compress:   s.rotationCompress,
		})
		if errSink == nil {
			errSink = rotationWriter
		} else {
			errSink = zap.CombineWriteSyncers(rotationWriter, errSink)
		}
	}
	var encoder zapcore.Encoder
	if s.encoder == "console" {
		config := zap.NewDevelopmentEncoderConfig()
		config.ConsoleSeparator = "|"
		encoder = zapcore.NewConsoleEncoder(config)
	} else {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	}

	if s.enableTraceId {
		encoder = NewTraceEncoder(encoder, s.tracer)
	}

	s.atomicLevel = zap.NewAtomicLevel()
	s.atomicLevel.SetLevel(parseLevel(s.level))

	core := zapcore.NewCore(
		encoder,
		sink,
		s.atomicLevel,
	)

	if errSink != nil {
		core = zapcore.NewTee(
			core,
			zapcore.NewCore(encoder, errSink, zap.NewAtomicLevelAt(zap.ErrorLevel)),
		)
	}

	var opts []Option
	if !s.disableStacktrace {
		opts = append(opts, zap.AddStacktrace(parseLevel(s.stackTraceLevel)))
	}

	if s.reportCaller {
		opts = append(opts, zap.AddCaller())
	}

	logger := zap.New(core, opts...)
	return logger, nil
}

type sugarProvider struct {
	gone.Flag

	zapLogger *zap.Logger `gone:"*"`
	wrapped   *wrappedLogger
}

func (s *sugarProvider) Provide(tagConf string) (Logger, error) {
	if s.wrapped == nil {
		s.wrapped = &wrappedLogger{Logger: s.zapLogger}
	}

	_, keys := gone.TagStringParse(tagConf)
	if len(keys) > 0 {
		if keys[0] != "" {
			return s.wrapped.Named(keys[0]), nil
		}
	}
	return s.wrapped, nil
}
