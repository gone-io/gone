package use_case

import (
	"github.com/gone-io/gone/v2"
	"testing"
)

func TestFuncInject(t *testing.T) {
	type Dep struct {
		gone.Flag
		ID int
	}

	fnExecuted := false
	fn := func(dep *Dep) {
		if dep.ID != 1 {
			t.Fatal("func inject failed")
		}
		fnExecuted = true
	}

	gone.
		NewApp().
		Load(&Dep{ID: 1}).
		Run(func(injector gone.FuncInjector) {
			wrapFunc, err := injector.InjectWrapFunc(fn, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			_ = wrapFunc()
		})

	if !fnExecuted {
		t.Fatal("func inject failed")
	}
}
