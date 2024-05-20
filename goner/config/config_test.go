package config

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
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

func Test_config_Suck(t *testing.T) {
	gone.Test(func(g *xGoner) {
		assert.Equal(t, 100, *g.x)
		assert.Equal(t, 100, g.xInt)
		assert.Equal(t, "100", g.xStr)
		assert.Equal(t, 100.0, g.xFloat)
		assert.Equal(t, uint(100), g.xUint)
		assert.Equal(t, int64(100), g.xInt64)
		assert.Equal(t, uint64(100), g.xUint64)

		assert.Equal(t, 100*time.Second, g.d)

		assert.Equal(t, 200, g.conf.Y)
		assert.Equal(t, 200, g.confP.Y)
		assert.Equal(t, "test", g.conf.X)
		assert.Equal(t, "test", g.confP.X)

		assert.Equal(t, 100, g.confL[0].Y)
		assert.Equal(t, 100, g.confL2[0].Y)
		assert.Equal(t, 200, g.confL[1].Y)
		assert.Equal(t, 200, g.confL2[1].Y)

		assert.Equal(t, "test1", g.confL[0].X)
		assert.Equal(t, "test1", g.confL2[0].X)
		assert.Equal(t, "test2", g.confL[1].X)
		assert.Equal(t, "test2", g.confL2[1].X)

	}, func(cemetery gone.Cemetery) error {
		cemetery.Bury(&xGoner{})
		return Priest(cemetery)
	})
}
