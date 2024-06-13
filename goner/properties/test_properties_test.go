package properties

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_getConfDir(t *testing.T) {
	_ = os.Setenv("CONF", "")
	dir := getConfDir()
	assert.Equal(t, "", dir)

	err := os.Setenv("CONF", "XXX")
	assert.Nil(t, err)
	dir = getConfDir()
	assert.Equal(t, "XXX", dir)

	x := "conf"
	confFlag = &x
	dir = getConfDir()
	assert.Equal(t, "conf", dir)

	err = os.Setenv("CONF", "")
	assert.Nil(t, err)
}

func Test_getExecutableDir(t *testing.T) {
	_, err := getExecutableDir()
	assert.Nil(t, err)
}
