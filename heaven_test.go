package gone_test

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"os"
	"syscall"
	"testing"
	"time"
)

type Triangle struct {
	gone.GonerFlag
	a gone.XPoint `gone:"point-a"`
	b gone.Point  `gone:"point-b"`
	c *gone.Point `gone:"point-c"`
}

func TestRun(t *testing.T) {
	go func() {
		time.Sleep(1 * time.Second)
		process, _ := os.FindProcess(os.Getpid())
		_ = process.Signal(syscall.SIGQUIT)
	}()
	gone.Run(func(cemetery gone.Cemetery) error {
		cemetery.
			Bury(&Triangle{}).
			Bury(&gone.Point{Index: 1}, "point-a").
			Bury(&gone.Point{Index: 2}, "point-b").
			Bury(&gone.Point{Index: 3}, "point-c")

		return nil
	})
}

func TestNew(t *testing.T) {
	var sort []int

	const gonerId = "Triangle"

	heaven := gone.
		New(func(cemetery gone.Cemetery) error {
			cemetery.
				Bury(&Triangle{}, gonerId).
				Bury(&gone.Point{Index: 1}, "point-a").
				Bury(&gone.Point{Index: 2}, "point-b").
				Bury(&gone.Point{Index: 3}, "point-c")

			return nil
		})

	heaven.
		BeforeStart(func(cemetery gone.Cemetery) error {
			tomb := cemetery.GetTomById(gonerId)
			triangle, ok := tomb.GetGoner().(*Triangle)
			assert.True(t, ok)
			assert.Equal(t, triangle.a.GetIndex(), 1)
			assert.Equal(t, triangle.b.GetIndex(), 2)
			assert.Equal(t, triangle.c.Index, 3)

			sort = append(sort, 0)
			return nil
		}).
		AfterStart(func(cemetery gone.Cemetery) error {
			sort = append(sort, 1)
			return nil
		}).
		BeforeStop(func(cemetery gone.Cemetery) error {
			sort = append(sort, 2)
			return nil
		}).
		AfterStop(func(cemetery gone.Cemetery) error {
			sort = append(sort, 3)
			return nil
		})

	go func() {
		time.Sleep(1 * time.Second)
		heaven.Stop()
	}()
	heaven.Start()

	for i := range sort {
		assert.Equal(t, i, sort[i])
	}
}
