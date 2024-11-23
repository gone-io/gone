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

type testBird struct {
	Name string
}

func (b *testBird) Fly() {
	println(b.Name + " flying")
}

type testCat struct {
	Flag
	Name string
}

func (*testCat) Meow() {
	println("meow")
}

func TestNewProvider(t *testing.T) {
	t.Run("provide struct", func(t *testing.T) {
		RunTest(func(p struct {
			blackBird testBird `gone:"*,black"`
			grayBird  testBird `gone:"*,gray"`
		}) {
			assert.Equal(t, p.blackBird.Name, "black")
			assert.Equal(t, p.grayBird.Name, "gray")
		}, func(cemetery Cemetery) error {
			cemetery.Bury(&testCat{
				Name: "cat",
			})

			priest := NewProviderPriest(func(tagConf string, in struct {
				cat *testCat `gone:"*"`
			}) (testBird, error) {
				assert.Equal(t, in.cat.Name, "cat")
				return testBird{Name: tagConf}, nil
			})
			return priest(cemetery)
		})
	})

	t.Run("provide struct pointer", func(t *testing.T) {
		RunTest(func(p struct {
			blackBird *testBird `gone:"*,black"`
			grayBird  *testBird `gone:"*,gray"`
		}) {
			assert.Equal(t, p.blackBird.Name, "black")
			assert.Equal(t, p.grayBird.Name, "gray")
		}, func(cemetery Cemetery) error {
			cemetery.Bury(&testCat{
				Name: "cat",
			})

			priest := NewProviderPriest(func(tagConf string, in struct {
				cat *testCat `gone:"*"`
			}) (*testBird, error) {
				assert.Equal(t, in.cat.Name, "cat")
				return &testBird{Name: tagConf}, nil
			})
			return priest(cemetery)
		})
	})
}
