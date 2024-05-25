package priest

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_priestParamReg(t *testing.T) {
	var s = "cemetery gone.Cemetery"
	match := priestParamReg.MatchString(s)
	assert.True(t, match)
}

func Test_goModuleInfo(t *testing.T) {
	t.Run("err1", func(t *testing.T) {
		_, _, err := goModuleInfo("testdata/x/sub")
		assert.Error(t, err)
	})

	t.Run("err2", func(t *testing.T) {
		_, _, err := goModuleInfo("testdata/x/sub2/.keep")
		assert.Error(t, err)
	})

	t.Run("err3", func(t *testing.T) {
		_ = os.MkdirAll("testdata/x/sub3/sub/sub", 0766)
		_, _, err := goModuleInfo("testdata/x/sub3")
		assert.Error(t, err)
	})
}
