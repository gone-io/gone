package gone

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

func TestGetDefaultLogger(t *testing.T) {
	logger := GetDefaultLogger()
	if logger == nil {
		t.Error("GetDefaultLogger() should not return nil")
	}
}

func TestLoggerLevels(t *testing.T) {
	logger := GetDefaultLogger()

	tests := []struct {
		level     LoggerLevel
		levelName string
	}{
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.levelName, func(t *testing.T) {
			logger.SetLevel(tt.level)
			if logger.GetLevel() != tt.level {
				t.Errorf("SetLevel(%v) failed, got %v", tt.level, logger.GetLevel())
			}
		})
	}
}

func TestLogOutput(t *testing.T) {
	// 捕获日志输出
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr) // 测试结束后恢复默认输出

	logger := GetDefaultLogger()

	tests := []struct {
		level     LoggerLevel
		logFunc   func(string, ...any)
		message   string
		shouldLog bool
	}{
		{InfoLevel, logger.Debugf, "debug message", false},
		{InfoLevel, logger.Infof, "info message", true},
		{InfoLevel, logger.Warnf, "warn message", true},
		{InfoLevel, logger.Errorf, "error message", true},

		{ErrorLevel, logger.Debugf, "debug message", false},
		{ErrorLevel, logger.Infof, "info message", false},
		{ErrorLevel, logger.Warnf, "warn message", false},
		{ErrorLevel, logger.Errorf, "error message", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.level), func(t *testing.T) {
			buf.Reset()
			logger.SetLevel(tt.level)
			tt.logFunc(tt.message)

			output := buf.String()
			hasMessage := strings.Contains(output, tt.message)

			if tt.shouldLog && !hasMessage {
				t.Errorf("expected message %q to be logged, but it wasn't", tt.message)
			}
			if !tt.shouldLog && hasMessage {
				t.Errorf("expected message %q not to be logged, but it was", tt.message)
			}
		})
	}
}

var strPointer = func(str string) *string {
	return &str
}

func Test_defaultLogger_Level(t *testing.T) {

	tests := []struct {
		name  string
		level *string
		want  LoggerLevel
	}{
		{
			name:  "debug",
			level: nil,
			want:  InfoLevel,
		},
		{
			name:  "info",
			level: new(string),
			want:  InfoLevel,
		},
		{
			name:  "warn",
			level: strPointer("warn"),
			want:  WarnLevel,
		},
		{
			name:  "error",
			level: strPointer("error"),
			want:  ErrorLevel,
		},
		{
			name:  "unknown",
			level: strPointer("unknown"),
			want:  InfoLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &defaultLogger{
				level: tt.level,
			}
			if got := l.Level(); got != tt.want {
				t.Errorf("Level() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultLogger_Debugf(t *testing.T) {
	type args struct {
		msg  string
		args []any
	}
	tests := []struct {
		name  string
		level *string
		args  args
	}{
		{
			name:  "debug",
			level: strPointer("debug"),
			args: args{
				msg:  "debug message",
				args: []any{"arg1", "arg2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &defaultLogger{
				level: tt.level,
			}
			l.Debugf(tt.args.msg, tt.args.args...)
		})
	}
}

func Test_defaultLogger_SetLevel(t *testing.T) {

	type args struct {
		level LoggerLevel
	}

	tests := []struct {
		name  string
		args  args
		check func(level string)
	}{
		{
			name: "debug",
			args: args{
				level: DebugLevel,
			},
			check: func(level string) {
				if level != "debug" {
					t.Errorf("SetLevel() = %v, want %v", level, "debug")
				}
			},
		},
		{
			name: "info",
			args: args{
				level: InfoLevel,
			},
			check: func(level string) {
				if level != "info" {
					t.Errorf("SetLevel() = %v, want %v", level, "info")
				}
			},
		},
		{
			name: "warn",
			args: args{
				level: WarnLevel,
			},
			check: func(level string) {
				if level != "warn" {
					t.Errorf("SetLevel() = %v, want %v", level, "warn")
				}
			},
		},
		{
			name: "error",
			args: args{
				level: ErrorLevel,
			},
			check: func(level string) {
				if level != "error" {
					t.Errorf("SetLevel() = %v, want %v", level, "error")
				}
			},
		},
		{
			name: "unknown",
			args: args{
				level: 0,
			},
			check: func(level string) {
				if level != "info" {
					t.Errorf("SetLevel() = %v, want %v", level, "info")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &defaultLogger{
				level: new(string),
			}
			l.SetLevel(tt.args.level)
			tt.check(*l.level)
		})
	}
}
