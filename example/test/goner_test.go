package test

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Line(t *testing.T) {
	t.Run("config default", func(t *testing.T) {
		gone.TestAt(pointNameA, func(point *Point) {
			assert.Equal(t, point.X, 1000)
			assert.Equal(t, point.Y, 200)
		}, config.Priest, Priest)
	})

	t.Run("config default", func(t *testing.T) {
		gone.Test(func(line *Line) {
			assert.Equal(t, line.A.Y, 200)
		}, Priest)
	})

	t.Run("ReplaceBury", func(t *testing.T) {
		gone.Test(func(line *Line) {
			assert.Equal(t, line.A.X, 20)
		}, Priest, func(cemetery gone.Cemetery) error {
			Mock := func() gone.Goner {
				return &Point{X: 20}
			}
			return cemetery.ReplaceBury(Mock(), pointNameA)
		})
	})
}
