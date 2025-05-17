package use_case

import (
	"github.com/gone-io/gone/v2"
	"testing"
)

func TestReplaceAndDefault(t *testing.T) {
	type X struct {
		gone.Flag
		ID int
	}
	gone.
		NewApp().
		Load(&X{ID: 1}, gone.Name("x"), gone.IsDefault()).
		Load(&X{ID: 2}, gone.Name("x"), gone.ForceReplace()).
		Load(&X{ID: 3}, gone.IsDefault()).
		Run(func(xList []*X, x *X) {
			if len(xList) != 2 {
				t.Errorf("xList error")
			}
			if x.ID != 3 {
				t.Errorf("x error")
			}
		})

}
