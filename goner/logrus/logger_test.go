package logrus

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_parseOutput(t *testing.T) {
	gone.Prepare(Priest).Test(func(log gone.Logger) {
		log.Info("info log")
		log.Warn("warn log")
		log.Error("error log")
	})
}

func Test_parseLogLevel(t *testing.T) {
	defer func() {
		err := recover()
		assert.Nil(t, err)
	}()
	parseLogLevel("xxx")
}

func Test_parseOutput1(t *testing.T) {
	t.Run("create log file failed", func(t *testing.T) {
		defer func() {
			err := recover()
			assert.Error(t, err.(gone.Error))
		}()
		_ = parseOutput("testdata/noop/test.log")
	})

	t.Run("create log file success", func(t *testing.T) {
		f := parseOutput("testdata/log/test.log")
		assert.NotNil(t, f)
		defer f.(*os.File).Close()
	})
	t.Run("stderr", func(t *testing.T) {
		f := parseOutput("stderr")

		assert.Equal(t, os.Stderr, f)
	})
	t.Run("stdout", func(t *testing.T) {
		f := parseOutput("stdout")

		assert.Equal(t, os.Stdout, f)
	})
}
