package gone

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type errProphet struct {
	Flag
}

func (e *errProphet) AfterRevive() error {
	return errors.New("AfterReviveError")
}

func Test_Test(t *testing.T) {
	t.Run("suc", func(t *testing.T) {
		type Line struct {
			Flag
			A XPoint `gone:"point-a"`
			b XPoint `gone:"point-b"`
		}

		var a = &Point{x: 1}
		var b = &Point{x: 2}

		var executed = false
		Test(func(l *Line) {
			assert.Equal(t, a, l.A)
			assert.Equal(t, b, l.b)

			executed = true
		}, func(cemetery Cemetery) error {
			cemetery.Bury(a, GonerId("point-a"))
			cemetery.Bury(b, GonerId("point-b"))
			cemetery.Bury(&Line{})
			return nil
		})
		assert.True(t, executed)
	})

	t.Run("failed: CannotFoundGonerById", func(t *testing.T) {
		var executed = false
		func() {
			defer func() {
				a := recover()
				assert.Equal(t, CannotFoundGonerById, a.(Error).Code())
				executed = true
			}()
			TestAt("point-a", func(p *Point) {

			}, func(cemetery Cemetery) error {
				return nil
			})
		}()

		assert.True(t, executed)
	})

	t.Run("failed: CannotFoundGonerByType", func(t *testing.T) {
		var executed = false
		func() {
			defer func() {
				a := recover()
				assert.Equal(t, CannotFoundGonerByType, a.(Error).Code())
				executed = true
			}()
			Test(func(p *Point) {

			}, func(cemetery Cemetery) error {
				return nil
			})
		}()

		assert.True(t, executed)
	})

	t.Run("failed: AfterRevive err", func(t *testing.T) {
		var executed = false

		func() {
			defer func() {
				a := recover()
				assert.Equal(t, "AfterReviveError", a.(error).Error())
				executed = true
			}()
			Test(func(p *errProphet) {

			}, func(cemetery Cemetery) error {
				cemetery.Bury(&errProphet{})
				return nil
			})
		}()

		assert.True(t, executed)
	})
}

func Test_TestAt(t *testing.T) {
	t.Run("suc", func(t *testing.T) {
		var executed = false
		type Line struct {
			Flag
			A XPoint `gone:"point-a"`
			b XPoint `gone:"point-b"`
		}

		var a = &Point{x: 1}
		var b = &Point{x: 2}

		TestAt("point-a", func(p *Point) {
			assert.Equal(t, p, a)

			executed = true
		}, func(cemetery Cemetery) error {
			cemetery.Bury(a, GonerId("point-a"))
			cemetery.Bury(b, GonerId("point-b"))
			cemetery.Bury(&Line{})
			return nil
		})
		assert.True(t, executed)
	})

	t.Run("suc: more than one Goner found by type", func(t *testing.T) {
		var executed = false
		a := &Point{}
		b := &Point{}

		Test(func(p *Point) {
			executed = true
			assert.Equal(t, a, p)
		}, func(cemetery Cemetery) error {
			cemetery.Bury(a)
			cemetery.Bury(b)
			return nil
		})

		assert.True(t, executed)
	})

	t.Run("failed: NotCompatible", func(t *testing.T) {
		var executed = false
		func() {
			defer func() {
				a := recover()
				assert.Equal(t, NotCompatible, a.(Error).Code())
				executed = true
			}()

			type Line struct {
				Flag
			}
			TestAt("point-a", func(p *Point) {

			}, func(cemetery Cemetery) error {
				cemetery.Bury(&Line{}, GonerId("point-a"))
				return nil
			})
		}()

		assert.True(t, executed)
	})

	t.Run("failed: NotCompatible", func(t *testing.T) {
		var executed = false
		func() {
			defer func() {
				a := recover()
				assert.Equal(t, NotCompatible, a.(Error).Code())
				executed = true
			}()

			type Line struct {
				Flag
			}
			TestAt("point-a", func(p *Point) {

			}, func(cemetery Cemetery) error {

				cemetery.Bury(&Line{}, GonerId("point-a"))
				return nil
			})
		}()

		assert.True(t, executed)
	})
}

type angel struct {
	Flag
	x int
}

func (i *angel) Start(Cemetery) error {
	i.x = 100
	return nil
}

func (i *angel) Stop(Cemetery) error {
	return nil
}

func (i *angel) X() int {
	return i.x
}

func Test_testHeaven_installAngelHook(t *testing.T) {
	type UseAngel struct {
		Flag
		angel *angel `gone:"*"`
	}
	var executed = false
	Test(func(u *UseAngel) {
		assert.Equal(t, 100, u.angel.X())
		executed = true
	}, func(cemetery Cemetery) error {
		cemetery.Bury(&angel{})
		cemetery.Bury(&UseAngel{})
		return nil
	})
	assert.True(t, executed)
}

func TestPreparer_Test(t *testing.T) {
	Prepare().Test(func(in struct {
		cemetery Cemetery `gone:"gone-cemetery"`
	}) {
		assert.NotNil(t, in.cemetery)
	})
}

func TestBuryMockCemetery_Bury(t *testing.T) {
	cemetery := NewBuryMockCemeteryForTest()
	point, id := &Point{}, "point-x"
	cemetery.Bury(point, GonerId(id))

	cemetery.Bury(&Point{x: 100})

	tomb := cemetery.GetTomById(GonerId(id))
	assert.Equal(t, point, tomb.GetGoner())

	tombs := cemetery.GetTomByType(reflect.TypeOf(*point))
	assert.Equal(t, 2, len(tombs))
}
