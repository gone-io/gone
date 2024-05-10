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

func forText(in struct {
	a Point `gone:"point-a"`
	b Point `gone:"point-b"`
}) int {
	println(in.a.GetIndex(), in.b.GetIndex())
	return in.a.GetIndex() + in.b.GetIndex()
}

func TestInjectWrapFn(t *testing.T) {
	heaven :=
		New(func(cemetery Cemetery) error {
			cemetery.
				Bury(&Point{Index: 1}, "point-a").
				Bury(&Point{Index: 2}, "point-b").
				Bury(&Point{Index: 3}, "point-c")

			return nil
		})

	flag := 0
	heaven.AfterStart(func(cemetery Cemetery) error {
		fn, err := InjectWrapFn(cemetery, forText)
		assert.Nil(t, err)

		results := ExecuteInjectWrapFn(fn)
		assert.Equal(t, 3, results[0])

		flag = 1

		return nil
	})
	heaven.Install()
	heaven.Start()
	assert.Equal(t, 1, flag)
}
