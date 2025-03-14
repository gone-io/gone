package use_case

import (
	"github.com/gone-io/gone/v2"
	"testing"
)

func TestAllowNil(t *testing.T) {
	type Dep struct {
		gone.Flag
	}

	type Dep2 struct {
		gone.Flag
	}

	type AllowNil struct {
		gone.Flag
		dep *Dep `gone:"*" option:"allowNil"`
	}

	provider := gone.WrapFunctionProvider(func(extend string, in struct {
		core *gone.Core `gone:"*"`
	}) (*Dep, error) {
		if extend == "ok" {
			return &Dep{}, nil
		} else {
			return nil, gone.NewInnerError("err", 0)
		}
	})

	t.Run("use allowNil", func(t *testing.T) {
		gone.
			NewApp().
			Load(&AllowNil{}).
			Load(&Dep2{}, gone.Name("dep2")).
			Load(provider, gone.Name("p")).
			Run(func(in struct {
				dep  *Dep  `gone:"*" option:"allowNil"`
				dep2 *Dep2 `gone:"dep2"`
				dep3 *Dep2 `gone:"dep3" option:"allowNil"`
				dep4 *Dep  `gone:"p,ok"`
				dep5 *Dep  `gone:"p,err" option:"allowNil"`
			}) {
				if in.dep != nil {
					t.Error("dep should be nil")
				}
				if in.dep2 == nil {
					t.Error("dep2 should not be nil")
				}
				if in.dep3 != nil {
					t.Error("dep3 should be nil")
				}
				if in.dep4 == nil {
					t.Error("dep4 should not be nil")
				}
				if in.dep5 != nil {
					t.Error("dep5 should be nil")
				}
			})
	})

	t.Run("not use and panic", func(t *testing.T) {
		type TestCase struct {
			name string
			fn   any
		}

		testCases := []TestCase{
			{
				name: "inject by type",
				fn: func(in struct {
					dep *Dep `gone:"*"`
				}) {
				},
			},
			{
				name: "inject by name",
				fn: func(in struct {
					dep3 *Dep2 `gone:"dep3"`
				}) {
				},
			},
			{
				name: "inject by provider name",
				fn: func(in struct {
					dep5 *Dep `gone:"p,err"`
				}) {
				},
			},
		}

		for _, ca := range testCases {
			t.Run(ca.name, func(t *testing.T) {
				defer func() {
					if err := recover(); err == nil {
						t.Error("should panic")
					}
				}()

				gone.
					NewApp().
					Load(&AllowNil{}).
					Load(&Dep2{}, gone.Name("dep2")).
					Load(provider, gone.Name("p")).
					Run(ca.fn)

				t.Error("should panic")
			})
		}

	})

}
