package priest

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_priestParamReg(t *testing.T) {
	var s = "cemetery gone.Cemetery"
	match := priestParamReg.MatchString(s)
	assert.True(t, match)
}
