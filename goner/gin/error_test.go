package gin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBusinessError(t *testing.T) {
	businessError := NewBusinessError("error", 100)
	assert.Equal(t, "error", businessError.Msg())
	assert.Equal(t, 100, businessError.Code())
}
