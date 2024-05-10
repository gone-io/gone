package gone_test

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrepare(t *testing.T) {
	i := 0
	gone.
		Prepare(func(cemetery gone.Cemetery) error {
			cemetery.
				Bury(&Triangle{}).
				Bury(&gone.Point{Index: 1}, "point-a").
				Bury(&gone.Point{Index: 2}, "point-b").
				Bury(&gone.Point{Index: 3}, "point-c")

			return nil
		}).
		BeforeStart(func(in struct {
			a      gone.Point   `gone:"point-a"`
			points []gone.Point `gone:"*"`
		}) {
			i++
			assert.Equal(t, 2, i)
			assert.Equal(t, in.a.GetIndex(), 1)
			assert.Equal(t, 3, len(in.points))
		}).
		BeforeStop(func() {
			i++
			assert.Equal(t, 4, i)
		}).
		AfterStart(func() {
			i++
			assert.Equal(t, 3, i)
		}).
		BeforeStart(func() {
			i++
			assert.Equal(t, 1, i)
		}).
		Run()

	assert.Equal(t, 4, i)
}
