package mock

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_isInputFromPipe(t *testing.T) {
	assert.False(t, isInputFromPipe())
}

func Test_getFile(t *testing.T) {
	t.Run("filepath is empty", func(t *testing.T) {
		file, err := getFile("")
		assert.Error(t, err)
		assert.Nil(t, file)
	})

	t.Run("file existed", func(t *testing.T) {
		file, err := getFile("testdata/x-testInterface.go")
		assert.Error(t, err)
		assert.Nil(t, file)
	})

	t.Run("file is dir", func(t *testing.T) {
		file, err := getFile("testdata")
		assert.Error(t, err)
		assert.Nil(t, file)
	})
}
