package properties_test

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

type Conf struct {
	X string `properties:"x"`
	Y int    `properties:"y"`
}

type xGoner struct {
	gone.Flag
	x       *int    `gone:"config,test.x"`
	xInt    int     `gone:"config,test.x"`
	xStr    string  `gone:"config,test.x"`
	xFloat  float64 `gone:"config,test.x"`
	xUint   uint    `gone:"config,test.x"`
	xInt64  int64   `gone:"config,test.x"`
	xUint64 uint64  `gone:"config,test.x"`

	d time.Duration `gone:"config,test.d"`

	conf   Conf    `gone:"config,test.conf"`
	confP  *Conf   `gone:"config,test.conf"`
	confL  []Conf  `gone:"config,test.list.conf"`
	confL2 []*Conf `gone:"config,test.list.conf"`
}

func TestGetProperties(t *testing.T) {

	err := os.Setenv("ENV", "test")
	assert.Nil(t, err)

	err = os.Setenv("CONF", "test")
	assert.Nil(t, err)

	gone.Prepare(func(cemetery gone.Cemetery) error {
		cemetery.Bury(&xGoner{})
		return config.Priest(cemetery)
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
