package gone

import (
	"github.com/stretchr/testify/assert"
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
			assert.Equal(t, test, test2)
		})
}
