package use_case

import (
	"github.com/gone-io/gone"
	"testing"
)

type hookTest struct {
	gone.Flag
	beforeStart gone.BeforeStart `gone:""`
	afterStart  gone.AfterStart  `gone:""`
	beforeStop  gone.BeforeStop  `gone:""`
	afterStop   gone.AfterStop   `gone:""`
}

var orders []int

func (h *hookTest) Init() {
	h.beforeStart(func() {
		println("before start 1")
		orders = append(orders, 1)
	})

	h.beforeStart(func() {
		println("before start 2")
		orders = append(orders, 2)
	})

	h.afterStart(func() {
		println("after start 3")
		orders = append(orders, 3)
	})

	h.afterStart(func() {
		println("after start 4")
		orders = append(orders, 4)
	})

	h.beforeStop(func() {
		println("before stop 5")
		orders = append(orders, 5)
	})

	h.beforeStop(func() {
		println("before stop 6")
		orders = append(orders, 6)
	})

	h.afterStop(func() {
		println("after stop 7")
		orders = append(orders, 7)
	})
	h.afterStop(func() {
		println("after stop 8")
		orders = append(orders, 8)
	})
}

func TestUseHook(t *testing.T) {
	gone.Load(&hookTest{}).Run()

	wantedOrder := []int{2, 1, 3, 4, 6, 5, 7, 8}
	for i := range wantedOrder {
		if wantedOrder[i] != orders[i] {
			t.Errorf("wanted %v, got %v", wantedOrder[i], orders[i])
		}
	}
}

func TestUseHookDirectly(t *testing.T) {
	type testGoner struct {
		gone.Flag
	}
	gone.
		Load(&testGoner{}).
		BeforeStart(func() {
			println(" BeforeStart")
		}).
		AfterStart(func() {
			println(" AfterStart")
		}).
		Run()
}
