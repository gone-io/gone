package gone

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_cemetery_reviveOne(t *testing.T) {
	type A struct {
		DeadFlag
		a int
		b int
	}

	type B struct {
		DeadFlag
		a A `gone:"*"`
		A A `gone:"*"`
	}

	cemetery := NewCemetery()

	cemetery.Bury(&A{}, "")
	cemetery.Bury(&B{}, "")

	err := cemetery.revive()
	assert.Nil(t, err)
}

type A struct {
	DeadFlag
	a int `gone:"xxxx,zzz,xxx,yyy"`
	b int
}

type B struct {
	DeadFlag
	a A `gone:"*"`
	A A `gone:"*"`
}
