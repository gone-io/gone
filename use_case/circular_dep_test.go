package use_case

import (
	"github.com/gone-io/gone/v2"
	"strings"
	"testing"
)

type depA1 struct {
	gone.Flag
	dep *depB1 `gone:"*"`
}

type depB1 struct {
	gone.Flag
	dep *depA1 `gone:"*"`
}

func TestCircularDependency1(t *testing.T) {
	gone.
		NewApp().
		Load(&depA1{}).
		Load(&depB1{}).
		Run()
}

type depA2 struct {
	gone.Flag
	dep *depB2 `gone:"*"`
}

func (d *depA2) Init() {}

type depB2 struct {
	gone.Flag
	dep *depA2 `gone:"*"`
}

func (d *depB2) Init() {}

func TestCircularDependency2(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if !strings.Contains(r.(error).Error(), "circular dependency") {
				t.Errorf("Expected panic with circular dependency error, got: %v", r)
			}
		}
	}()
	gone.
		NewApp().
		Load(&depA2{}).
		Load(&depB2{}).
		Run()
}

type depA3 struct {
	gone.Flag
	dep *depB3 `gone:"*"`
}
type depB3 struct {
	gone.Flag
	dep *depA3 `gone:"*"`
}

func (d *depB3) Init() {}

func TestCircularDependency3(t *testing.T) {
	gone.
		NewApp().
		Load(&depA3{}).
		Load(&depB3{}).
		Run()
}
