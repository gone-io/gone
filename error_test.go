package gone

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_iError_Error(t *testing.T) {
	innerError := NewInnerError("test", 100)
	s := innerError.Error()

	assert.True(t, strings.Contains(s, "Test_iError_Error"))
}
