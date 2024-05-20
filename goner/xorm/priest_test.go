package xorm

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPriest(t *testing.T) {
	testCemetery := gone.NewBuryMockCemeteryForTest()
	err := Priest(testCemetery)
	assert.Nil(t, err)
	err = Priest(testCemetery)
	assert.Nil(t, err)
}
