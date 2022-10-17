package gone_test

import (
	"github.com/gone-io/gone"
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
		_ = process.Signal(syscall.SIGINT)
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
