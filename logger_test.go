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
