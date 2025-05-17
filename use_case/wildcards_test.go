package use_case

import (
	"github.com/gone-io/gone/v2"
	"testing"
)

func TestUseWildcards(t *testing.T) {
	type X struct {
		gone.Flag
		ID int
	}

	xProvider := gone.WrapFunctionProvider(func(tagConf string, in struct{}) (*X, error) {
		return &X{}, nil
	})

	t.Run("inject to slice", func(t *testing.T) {
		type Y struct {
			gone.Flag
			t1 []*X `gone:"*"`
			t2 []*X `gone:"test-*"`
			t3 []*X `gone:"test-1*"`
			t4 []*X `gone:"test-?234"`
		}

		gone.
			NewApp().
			Load(&X{ID: 1}).
			Load(&X{ID: 2}, gone.Name("test-2")).
			Load(&X{ID: 3}, gone.Name("test-12")).
			Load(&X{ID: 4}, gone.Name("test-123")).
			Load(xProvider, gone.Name("test-1234")).
			Load(xProvider, gone.Name("test-2234")).
			Load(&Y{}).
			Run(func(y *Y) {
				if len(y.t1) != 6 {
					t.Fatal("inject to slice failed")
				}
				if len(y.t2) != 5 {
					t.Fatal("inject to slice failed")
				}
				if len(y.t3) != 3 {
					t.Fatal("inject to slice failed")
				}
				if len(y.t4) != 2 {
					t.Fatal("inject to slice failed")
				}
			})
	})

	t.Run("inject to pointer or interface", func(t *testing.T) {
		gone.
			NewApp().
			Load(&X{ID: 1}).
			Load(&X{ID: 2}, gone.Name("test-2")).
			Load(&X{ID: 3}, gone.Name("test-12")).
			Load(&X{ID: 4}, gone.Name("test-123")).
			Load(xProvider, gone.Name("test-1234")).
			Load(xProvider, gone.Name("test-2234")).
			Run(func(y struct {
				x  *X  `gone:"test-?23*"`
				x1 any `gone:"test-?23*"`
				x2 *X  `gone:"test-123"` // not use wildcards
			}) {
				if y.x == nil {
					t.Fatal("inject to pointer failed")
				}
				if y.x.ID != 4 {
					t.Fatal("inject to pointer failed")
				}
				if y.x1 != y.x {
					t.Fatal("inject to interface failed")
				}

				if y.x2 != y.x {
					t.Fatal("inject to pointer failed")
				}
			})
	})
}
