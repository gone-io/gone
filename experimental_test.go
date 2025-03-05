package gone

import (
	"testing"
)

func TestWrapFunctionProvider(t *testing.T) {
	type Test struct {
	}

	var test Test

	provider := WrapFunctionProvider(func(tagConf string, in struct{}) (Test, error) {
		return test, nil
	})

	Prepare(func(loader Loader) error {
		return loader.Load(provider)
	}).
		Test(func(test2 Test) {
			if test != test2 {
				t.Errorf("Expected %v, got %v", test, test2)
			}
		})
}
