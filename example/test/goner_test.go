package test

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Line(t *testing.T) {
	t.Run("config default", func(t *testing.T) {
		gone.
			TestKit(&Point{}, func(cemetery gone.Cemetery) error {
				return config.Priest(cemetery)
			}).
			Run(func(point *Point) {
				assert.Equal(t, point.X, 1000)
				assert.Equal(t, point.Y, 200)
			})
	})

	t.Run("config default", func(t *testing.T) {
		gone.
			TestKit(&Line{}, Priest).
			Run(func(line *Line) {
				assert.Equal(t, line.A.Y, 200)
			})
	})

	t.Run("ReplaceBury", func(t *testing.T) {
		gone.
			TestKit(&Line{}, Priest, func(cemetery gone.Cemetery) error {
				Mock := func() gone.Goner {
					return &Point{X: 20}
				}
				return cemetery.ReplaceBury(Mock(), pointNameA)
			}).
			Run(func(line *Line) {
				assert.Equal(t, line.A.X, 20)
			})
	})
}
