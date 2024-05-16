package gone

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Test(t *testing.T) {
	type Line struct {
		Flag
		A XPoint `gone:"point-a"`
		b XPoint `gone:"point-b"`
	}

	var a = &Point{x: 1}
	var b = &Point{x: 2}

	var executed = false
	Test(func(l *Line) {
		assert.Equal(t, a, l.A)
		assert.Equal(t, b, l.b)

		executed = true
	}, func(cemetery Cemetery) error {
		cemetery.Bury(a, "point-a")
		cemetery.Bury(b, "point-b")
		cemetery.Bury(&Line{})
		return nil
	})
	assert.True(t, executed)
}

func Test_TestAt(t *testing.T) {
	t.Run("suc", func(t *testing.T) {
		var executed = false
		type Line struct {
			Flag
			A XPoint `gone:"point-a"`
			b XPoint `gone:"point-b"`
		}

		var a = &Point{x: 1}
		var b = &Point{x: 2}

		TestAt("point-a", func(p *Point) {
			assert.Equal(t, p, a)

			executed = true
		}, func(cemetery Cemetery) error {
			cemetery.Bury(a, "point-a")
			cemetery.Bury(b, "point-b")
			cemetery.Bury(&Line{})
			return nil
		})
		assert.True(t, executed)
	})
}

func Test_testHeaven_WithId(t *testing.T) {
	test := &testHeaven[*Point]{}
	result := test.WithId("point-a")
	assert.Equal(t, test, result)

	assert.Equal(t, "point-a", string(test.testGonerId))
}

type angel struct {
	Flag
	x int
}

func (i *angel) Start(Cemetery) error {
	i.x = 100
	return nil
}

func (i *angel) Stop(Cemetery) error {
	return nil
}

func (i *angel) X() int {
	return i.x
}

func Test_testHeaven_installAngelHook(t *testing.T) {
	type UseAngel struct {
		Flag
		angel *angel `gone:"*"`
	}
	var executed = false
	Test(func(u *UseAngel) {
		assert.Equal(t, 100, u.angel.X())
		executed = true
	}, func(cemetery Cemetery) error {
		cemetery.Bury(&angel{})
		cemetery.Bury(&UseAngel{})
		return nil
	})
	assert.True(t, executed)
}
