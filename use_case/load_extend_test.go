package use_case

import (
	"github.com/gone-io/gone/v2"
	"testing"
)

func TestLoadFunc(t *testing.T) {
	type X struct {
		gone.Flag
		ID int
	}

	var loadFunc = func(loader gone.Loader) error {
		return loader.Load(&X{ID: 1})
	}

	gone.
		NewApp(func(loader gone.Loader) error {
			loader.MustLoadX(loadFunc)
			loader.MustLoadX(loadFunc)
			loader.MustLoadX(loadFunc)
			loader.MustLoadX(loadFunc)
			return nil
		}).
		Run(func(x []*X) {
			if len(x) != 1 {
				t.Fatal("len(x) != 1")
			} else if x[0].ID != 1 {
				t.Fatal("x[0].ID != 1")
			}
		})
}
