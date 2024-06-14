package gone_zap

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"testing"
)

func Test_log_Suck(t *testing.T) {
	gone.Prepare(Priest).Test(func(in struct {
		log   Logger      `gone:"zap,in-test"`
		sugar gone.Logger `gone:"zap,sugar"`
	}) {

		in.log.Info("info log")
		in.sugar.Tracef("this is trace log")
	})
}

func Test_parseLevel(t *testing.T) {
	type args struct {
		level string
	}
	tests := []struct {
		name string
		args args
		want zapcore.Level
	}{
		{
			name: "debug",
			args: args{
				level: "debug",
			},
			want: zapcore.DebugLevel,
		},
		{
			name: "info",
			args: args{
				level: "info",
			},
			want: zapcore.InfoLevel,
		},
		{
			name: "warn",
			args: args{
				level: "warn",
			},
			want: zapcore.WarnLevel,
		},
		{
			name: "error",
			args: args{
				level: "error",
			},
			want: zapcore.ErrorLevel,
		},
		{
			name: "panic",
			args: args{
				level: "panic",
			},
			want: zapcore.PanicLevel,
		},
		{
			name: "fatal",
			args: args{
				level: "fatal",
			},
			want: zapcore.FatalLevel,
		},
		{
			name: "unknown",
			args: args{
				level: "unknown",
			},
			want: zapcore.InfoLevel, // default
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseLevel(tt.args.level); got != tt.want {
				t.Errorf("parseLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_log_Named(t *testing.T) {
	gone.Prepare(Priest).Test(func(l Logger) {
		named := l.Named("")
		assert.Equal(t, named, l)

		logger := l.Named("cat")

		assert.Equal(t, "cat", logger.(*log).Logger.Name())
	})
}

func Test_log_WithOptions(t *testing.T) {
	gone.Prepare(Priest).Test(func(l Logger) {
		assert.Equal(t, l, l.WithOptions())

		logger := l.WithOptions(zap.AddCallerSkip(1))
		assert.NotEqual(t, l, logger)
	})
}

func Test_log_With(t *testing.T) {
	gone.Prepare(Priest).Test(func(l Logger) {
		assert.Equal(t, l, l.With())

		logger := l.With(zap.String("key", "value"))
		assert.NotEqual(t, l, logger)
	})
}

func Test_log_Sugar(t *testing.T) {
	gone.Prepare(Priest).Test(func(l Logger) {
		l.Sugar().Infof("this is test:%d", 100)
	})
}

func Test_log_Build(t *testing.T) {
	_ = os.Setenv("ENV", "prod")
	gone.Prepare(Priest).Test(func(l Logger) {
		l.Info("info log")
		l.Error("error log")
	})
	_ = os.Setenv("ENV", "")
}
