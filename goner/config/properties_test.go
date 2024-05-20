package config

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestGetProperties(t *testing.T) {
	_, err := GetProperties()
	assert.Nil(t, err)

	err = os.Setenv("ENV", "test")
	assert.Nil(t, err)

	err = os.Setenv("CONF", "test")
	assert.Nil(t, err)

	gone.Prepare(func(cemetery gone.Cemetery) error {
		cemetery.Bury(&xGoner{})
		return Priest(cemetery)
	}).AfterStart(func(in struct {
		g *xGoner `gone:"*"`
	}) {
		assert.Equal(t, 500, *in.g.x)
		assert.Equal(t, 500, in.g.xInt)
		assert.Equal(t, "500", in.g.xStr)
		assert.Equal(t, 500.0, in.g.xFloat)
		assert.Equal(t, uint(500), in.g.xUint)
		assert.Equal(t, int64(500), in.g.xInt64)
		assert.Equal(t, uint64(500), in.g.xUint64)

		assert.Equal(t, 100*time.Second, in.g.d)
	}).Run()
}

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
