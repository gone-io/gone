package gone

import (
	"testing"
)

func TestWrapFunctionProvider(t *testing.T) {
	type Test struct {
	}

	var test Test

	t.Run("success", func(t *testing.T) {
		provider := WrapFunctionProvider(func(tagConf string, in struct{}) (Test, error) {
			return test, nil
		})

		NewApp(func(loader Loader) error {
			return loader.Load(provider)
		}).
			Test(func(test2 Test) {
				if test != test2 {
					t.Errorf("Expected %v, got %v", test, test2)
				}
			})
	})

	t.Run("inject err", func(t *testing.T) {
		type TestX struct {
		}

		provider := WrapFunctionProvider(func(tagConf string, in struct {
			Test *TestX `gone:"*"`
		}) (Test, error) {
			return test, ToError("test error")
		})

		err := SafeExecute(func() error {
			NewApp().
				Load(provider).
				Test(func(core Test) {
				})
			return nil
		})

		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("run err", func(t *testing.T) {
		provider := WrapFunctionProvider(func(tagConf string, in struct{}) (Test, error) {
			return test, ToError("test error")
		})

		err := SafeExecute(func() error {
			NewApp().
				Load(provider).
				Test(func(core Test) {
				})
			return nil
		})

		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}
