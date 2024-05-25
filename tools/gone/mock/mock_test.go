package mock

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_isInputFromPipe(t *testing.T) {
	assert.False(t, isInputFromPipe())
}
