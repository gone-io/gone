package gone

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type InjectedByProvider struct {
	Conf string
}

type TestInjectByProvider struct {
	Flag
}

func (*TestInjectByProvider) Suck(conf string, v reflect.Value, field reflect.StructField) error {
	provider := InjectedByProvider{
		Conf: conf,
	}
	v.Set(reflect.ValueOf(&provider))
	return nil
}

func NewTestInjectByProvider() (Goner, GonerOption) {
	return &TestInjectByProvider{}, Provide(&InjectedByProvider{})
}

func TestTestInjectByProvider(t *testing.T) {
	RunTest(func(dep struct {
		c1 *InjectedByProvider `gone:"*"`
		c2 *InjectedByProvider `gone:"*,conf-xxx,ok"`
	}) {
		assert.Equal(t, dep.c1.Conf, "")
		assert.Equal(t, dep.c2.Conf, "conf-xxx,ok")

	}, func(cemetery Cemetery) error {
		cemetery.Bury(NewTestInjectByProvider())
		return nil
	})
}
