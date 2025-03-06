package use_case

import (
	"github.com/gone-io/gone/v2"
	"testing"
)

func TestInjectByStructPointer(t *testing.T) {
	type Dep struct {
		gone.Flag
		Name string
	}

	type Injector struct {
		gone.Flag

		Dep *Dep `gone:"*"`
	}

	var dep1 = &Dep{
		Name: "dep1",
	}
	var dep2 = &Dep{
		Name: "dep2",
	}

	t.Run("inject single one", func(t *testing.T) {
		gone.
			Prepare(func(loader gone.Loader) error {
				_ = loader.Load(dep1)
				return nil
			}).
			Run(func(injector Injector) {
				if injector.Dep != dep1 {
					t.Fatal("injector.Dep != dep")
				}
			})
	})
	t.Run("use first when multi", func(t *testing.T) {
		gone.
			Prepare(func(loader gone.Loader) error {
				_ = loader.Load(dep1)
				_ = loader.Load(dep2)
				return nil
			}).
			Run(func(injector Injector) {
				if injector.Dep != dep1 {
					t.Fatal("injector.Dep != dep")
				}
			})
	})

	t.Run("use option.IsDefault when multi", func(t *testing.T) {
		gone.
			Prepare(func(loader gone.Loader) error {
				_ = loader.Load(dep1)
				_ = loader.Load(dep2, gone.IsDefault())
				return nil
			}).
			Run(func(injector Injector) {
				if injector.Dep != dep2 {
					t.Fatal("injector.Dep != dep2")
				}
			})
	})

	t.Run("use gone flag", func(t *testing.T) {
		gone.
			Prepare(func(loader gone.Loader) error {
				return loader.Load(dep1)
			}).
			Run(func(in struct {
				dep1 *Dep `gone:""`
				dep2 *Dep `gone:"*"`
			}) {
				if in.dep1 != dep1 {
					t.Fatal("injector.dep1 != dep")
				}
				if in.dep2 != dep1 {
					t.Fatal("injector.dep2 != dep")
				}
			})
	})

	// t.Run("register goners", func(t *testing.T) {
	// 	gone.
	// 		Load(dep1).
	// 		Loads(func(loader gone.Loader) error {

	// 		})
	// })
}
