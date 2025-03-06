package use_case

import (
	"github.com/gone-io/gone"
	"testing"
)

type worker interface {
	Work()
}

type workerImpl struct {
	gone.Flag
	name string
}

func (w *workerImpl) Work() {
	println("worker", w.name, "work")
}

type workerImpl2 struct {
	gone.Flag
	name string
}

func (w *workerImpl2) Work() {
	println("worker", w.name, "work")
}

type factory struct {
	gone.Flag
	workers []worker `gone:"*"`
}

func TestUseSlice(t *testing.T) {
	gone.
		Prepare().
		Load(&factory{}, gone.Name("factory")).
		Load(&workerImpl{name: "worker1"}, gone.Name("worker1")).
		Load(&workerImpl2{name: "worker2"}, gone.Name("worker2")).
		Run(func(f *factory) {
			if len(f.workers) != 2 {
				t.Fatal("worker count is not 2")
			}
		})

}
