package gone_zap

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPriest(t *testing.T) {
	gone.Prepare(Priest).Test(func(cemetery gone.Cemetery) {
		err := Priest(cemetery)
		assert.Nil(t, err)
	})
}
