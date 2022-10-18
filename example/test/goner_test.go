package test

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Line(t *testing.T) {
	gone.
		TestKit(&Line{}, Digger, func(cemetery gone.Cemetery) error {
			Mock := func() gone.Goner {
				return &Point{X: 20}
			}
			cemetery.ReplaceBury(Mock(), pointNameA)
			return nil
		}).
		Run(func(line *Line) {

			assert.Equal(t, line.A.Y, float64(1))
		})

	//gone.
	//	TestKit(&Line{}, Digger).
	//	RunAtId("line-id", func(line *Line) {
	//
	//		assert.Equal(t, line.A.Y, float64(1))
	//	})
}
