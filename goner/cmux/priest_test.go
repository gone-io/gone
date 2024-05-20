package cmux

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPriest(t *testing.T) {
	cemetery := gone.NewBuryMockCemeteryForTest()
	err := Priest(cemetery)
	assert.Nil(t, err)
}
