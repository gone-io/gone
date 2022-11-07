package gone

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_FuncName(t *testing.T) {
	name := GetFuncName(Test_FuncName)
	println(name)
	assert.Equal(t, name, "github.com/gone-io/gone.Test_FuncName")
	fn := func() {}

	println(GetFuncName(fn))

	assert.Equal(t, GetFuncName(fn), "github.com/gone-io/gone.Test_FuncName.func1")
}

type XX interface {
	Get()
}

var XXPtr *XX
var XXType = reflect.TypeOf(XXPtr).Elem()

func Test_getInterfaceType(t *testing.T) {
	interfaceType := GetInterfaceType(new(XX))
	assert.Equal(t, interfaceType, XXType)
}
