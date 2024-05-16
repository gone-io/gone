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
