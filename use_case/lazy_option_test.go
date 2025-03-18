package use_case

import (
	"github.com/gone-io/gone/v2"
	"testing"
)

type depA4 struct {
	gone.Flag
	dep *depB4 `gone:"*"`
}

func (d *depA4) Init() {}

type depB4 struct {
	gone.Flag
	dep *depA4 `gone:"*" option:"lazy"`
}

func (d *depB4) Init() {}

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
