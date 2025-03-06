package use_case

import (
	"github.com/gone-io/gone"
	"testing"
)

type funcTest struct {
	gone.Flag
	injector gone.FuncInjector `gone:"*"`
}

func (f *funcTest) Test(t *testing.T) {

	fn := func(factory *factory) {
		if factory == nil {
			t.Fatal("factory is nil")
		}
	}

	wrapped, err := f.injector.InjectWrapFunc(fn, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = wrapped()

	fn2 := func(factory *factory, worker worker) {
		if factory == nil {
			t.Fatal("factory is nil")
		}
		if worker == nil {
			t.Fatal("worker is nil")
		}
	}

	wrapped2, err := f.injector.InjectWrapFunc(fn2, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = wrapped2()
}

func TestFuncInjector(t *testing.T) {
	gone.
		Prepare().
		Load(&funcTest{}).
		Load(&factory{}, gone.Name("factory")).
		Load(&workerImpl{name: "worker1"}, gone.Name("worker1")).
		Run(func(f *funcTest) {
			f.Test(t)
		})
}
