package gone

import (
	"errors"
	"reflect"
	"testing"
)

func Test_core_GetGonerByName(t *testing.T) {
	type X struct {
		Flag
		ID int
	}

	NewApp().
		Load(&X{ID: 1001}, Name("x")).
		Run(func(k GonerKeeper) {
			name := k.GetGonerByName("x")
			if name == nil {
				t.Error("name is nil")
				return
			}
			if x, ok := name.(*X); !ok {
				t.Error("name is not *X")
				return
			} else if x.ID != 1001 {
				t.Error("name.ID is not 1001")
				return
			}

			y := k.GetGonerByName("y")
			if y != nil {
				t.Error("y is not nil")
				return
			}
		})
}

func Test_core_GetGonerByType(t *testing.T) {
	type X struct {
		Flag
		ID int
	}

	t.Run("suc", func(t *testing.T) {
		NewApp().
			Load(&X{ID: 1001}, Name("x")).
			Load(WrapFunctionProvider(func(tagConf string, param struct{}) (*X, error) {
				return &X{ID: 1002}, nil
			})).
			Run(func(k GonerKeeper) {
				x := k.GetGonerByType(reflect.TypeOf(&X{}))
				if x == nil {
					t.Error("x is nil")
					return
				}
				if x.(*X).ID != 1001 {
					t.Error("x.ID is not 1001")
					return
				}
			})
	})

	t.Run("provide err", func(t *testing.T) {
		NewApp().
			Load(WrapFunctionProvider(func(tagConf string, param struct{}) (*X, error) {
				return nil, errors.New("provide err")
			})).
			Run(func(k GonerKeeper) {
				err := SafeExecute(func() error {
					_ = k.GetGonerByType(reflect.TypeOf(&X{}))
					return nil
				})
				if err == nil {
					t.Errorf("err is nil")
				}
			})
	})

	t.Run("not found", func(t *testing.T) {
		NewApp().
			Run(func(k GonerKeeper) {
				g := k.GetGonerByType(reflect.TypeOf(&X{}))
				if g != nil {
					t.Error("g is not nil")
				}
			})
	})

}

func Test_core_InjectStruct(t *testing.T) {
	type dep struct {
		Flag
	}

	type X struct {
		x *dep `gone:"*"`
	}

	t.Run("suc", func(t *testing.T) {
		d := &dep{}
		NewApp().
			Load(d).
			Run(func(i StructInjector) {
				var x = &X{}
				err := i.InjectStruct(x)
				if err != nil {
					t.Errorf("err is not nil")
					return
				}
				if x.x != d {
					t.Errorf("x.x is not d")
				}
			})
	})

	t.Run("goner is not ptr", func(t *testing.T) {
		NewApp().
			Run(func(i StructInjector) {
				err := i.InjectStruct("test")
				if err == nil {
					t.Errorf("err is nil")
				}
			})
	})
	t.Run("goner is not ptr to struct", func(t *testing.T) {
		var test string

		NewApp().
			Run(func(i StructInjector) {
				err := i.InjectStruct(&test)
				if err == nil {
					t.Errorf("err is nil")
				}
			})
	})
}
