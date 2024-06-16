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
			assert.Equal(t, point.X, 0)
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
			return cemetery.ReplaceBury(Mock(), gone.GonerId(pointNameA))
		})
	})

	t.Run("Prepare.Test", func(t *testing.T) {
		gone.
			Prepare(Priest).
			Test(func(
				line *Line, //注入gone框架中注册的类型

				in struct { //注入匿名结构体
					point *Point `gone:"example-test-point-a"`
				},
			) {
				assert.Equal(t, line.A.Y, 200)
				assert.Equal(t, in.point.Y, 200)
			})
	})
}
