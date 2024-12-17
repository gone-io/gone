package gone_zap

import (
	"github.com/gone-io/gone"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type sugar struct {
	gone.Flag
	*zap.SugaredLogger
	provider *zapLoggerProvider `gone:"*"`
}

func (l *sugar) Name() string {
	return "gone-logger"
}

func (l *sugar) Init() error {
	logger, err := l.provider.Provide("")
	if err != nil {
		return gone.ToError(err)
	}
	l.SugaredLogger = logger.Sugar()
	return nil
}
func (l *sugar) GetLevel() gone.LoggerLevel {
	switch l.SugaredLogger.Level() {
	case zap.DebugLevel:
		return gone.DebugLevel
	case zap.InfoLevel:
		return gone.InfoLevel
	case zap.WarnLevel:
		return gone.WarnLevel
	case zap.ErrorLevel:
		return gone.ErrorLevel
	default:
		if l.SugaredLogger.Level() > zap.ErrorLevel {
			return gone.ErrorLevel
		} else {
			return gone.DebugLevel
		}
	}
}

func (l *sugar) SetLevel(level gone.LoggerLevel) {
	var zapLevel zapcore.Level
	switch level {
	case gone.DebugLevel:
		zapLevel = zap.DebugLevel
	case gone.InfoLevel:
		zapLevel = zap.InfoLevel
	case gone.WarnLevel:
		zapLevel = zap.WarnLevel
	case gone.ErrorLevel:
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.InfoLevel
	}
	l.provider.SetLevel(zapLevel)
}
