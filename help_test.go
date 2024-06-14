package gone

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
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

func TestTimeStat(t *testing.T) {
	t.Run("with log", func(t *testing.T) {
		defer TimeStat("test", time.Now(), func(format string, args ...any) {
			assert.Equal(t, "test", args[0])
			assert.Equal(t, 4, len(args))
		})

	})
	t.Run("without log", func(t *testing.T) {
		defer TimeStat("test", time.Now())
	})
}

func TestWrapNormalFnToProcess(t *testing.T) {
	Prepare().Test(func(cemetery Cemetery) {
		fn := WrapNormalFnToProcess(func(in struct {
			in Logger `gone:"xxx"`
		}) {
		})

		err := fn(cemetery)
		assert.Error(t, err)

		process := WrapNormalFnToProcess(func() error {
			return errors.New("err")
		})

		err = process(cemetery)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "err")
	})
}

func TestBlank(t *testing.T) {
	flag := Flag{}
	flag.goneFlag()
	id := GonerId("")
	id.option()

	isDefault := IsDefault(true)
	isDefault.option()
	order := Order(1)
	order.option()
}
