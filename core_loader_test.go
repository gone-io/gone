package gone

import (
	"testing"
)

func TestCore_MustLoadX(t *testing.T) {
	t.Run("load goner", func(t *testing.T) {
		type g struct {
			Flag
			Name string
		}

		NewApp(func(loader Loader) error {
			loader.
				MustLoadX(&g{Name: "test"})
			return nil
		}).Run(func(g *g) {
			if g.Name != "test" {
				t.Fatal("load goner failed")
			}
		})
	})
	t.Run("load func", func(t *testing.T) {
		type g struct {
			Flag
			Name string
		}

		loadFunc := func(loader Loader) error {
			loader.
				MustLoadX(&g{Name: "test"})
			return nil
		}

		NewApp(func(loader Loader) error {
			loader.MustLoadX(loadFunc)
			return nil
		}).Run(func(g *g) {
			if g.Name != "test" {
				t.Fatal("load goner failed")
			}
		})
	})

	t.Run("load func twice", func(t *testing.T) {
		type g struct {
			Flag
			Name string
		}
		type X struct {
			Flag
			g *g `gone:"*"`
		}

		loadFunc := func(loader Loader) error {
			loader.
				MustLoadX(&g{Name: "test"})
			return nil
		}

		NewApp(func(loader Loader) error {
			loader.MustLoadX(loadFunc).MustLoadX(loadFunc)
			return nil
		}).Run(func(gList []*g, w X) {
			if len(gList) != 1 {
				t.Fatal("load duplicated")
			}
			if w.g != gList[0] {
				t.Fatal("load duplicated")
			}
		})
	})

	t.Run("load unsupported type", func(t *testing.T) {
		err := SafeExecute(func() error {
			NewApp(func(loader Loader) error {
				loader.MustLoadX(1)
				return nil
			}).Run(func() {
				t.Fatal("should not run")
			})
			return nil
		})
		if err == nil {
			t.Fatal("should error")
		}
	})

}
