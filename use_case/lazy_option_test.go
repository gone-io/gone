package use_case

import (
	"github.com/gone-io/gone/v2"
	"testing"
)

type depA4 struct {
	gone.Flag
	dep *depB4 `gone:"*"`
}

func (d *depA4) Init() {
	if d.dep == nil {
		panic("depB4.dep should not be nil")
	}
}

type depB4 struct {
	gone.Flag
	dep *depA4 `gone:"*" option:"lazy"`
}

func (d *depB4) Init() {
	if d.dep.dep != nil {
		panic("depB4.dep should be nil")
	}
}

func TestCircularDependency4(t *testing.T) {
	gone.
		NewApp().
		Load(&depA4{}).
		Load(&depB4{}).
		Run(func(a4 *depA4, b4 *depB4) {
			if a4.dep == nil {
				t.Error("a4.dep should be not nil")
			}
			if b4.dep == nil {
				t.Error("b4.dep should be not nil")
			}
		})
}

type depA5 struct {
	gone.Flag
	dep *depB5 `gone:"*"`
}

func (d *depA5) Init() {
	if d.dep != nil {
		panic("depB4.dep should be nil")
	}
}

type depB5 struct {
	gone.Flag
	dep *depA5 `gone:"*"`
}

func (d *depB5) Init() {
	if d.dep == nil {
		panic("depB4.dep should not be nil")
	}
}

func TestCircularDependency5(t *testing.T) {
	gone.
		NewApp().
		Load(&depB5{}).
		Load(&depA5{}, gone.LazyFill()).
		Run(func(a4 *depA5, b4 *depB5) {
			if a4.dep == nil {
				t.Error("a4.dep should be not nil")
			}
			if b4.dep == nil {
				t.Error("b4.dep should be not nil")
			}
		})
}
