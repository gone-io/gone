package gone_test

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPrepare(t *testing.T) {
	i := 0
	gone.
		Prepare(func(cemetery gone.Cemetery) error {
			cemetery.
				Bury(&Triangle{}).
				Bury(&gone.Point{Index: 1}, gone.GonerId("point-a")).
				Bury(&gone.Point{Index: 2}, gone.GonerId("point-b")).
				Bury(&gone.Point{Index: 3}, gone.GonerId("point-c"))

			return nil
		}).
		BeforeStart(func(in struct {
			a      gone.Point   `gone:"point-a"`
			points []gone.Point `gone:"*"`
		}) {
			i++
			assert.Equal(t, 2, i)
			assert.Equal(t, in.a.GetIndex(), 1)
			assert.Equal(t, 3, len(in.points))
		}).
		BeforeStop(func() {
			i++
			assert.Equal(t, 4, i)
		}).
		AfterStart(func() {
			i++
			assert.Equal(t, 3, i)
		}).
		BeforeStart(func() {
			i++
			assert.Equal(t, 1, i)
		}).
		AfterStop(func() {
			i++
			assert.Equal(t, 5, i)
		}).
		Run()

	assert.Equal(t, 5, i)
}

func TestPreparer_Serve(t *testing.T) {
	i := 0
	executed := false
	gone.
		Prepare().
		AfterStart(func(in struct {
			h gone.Heaven `gone:"gone-heaven"`
		}) {
			signal := in.h.GetHeavenStopSignal()

			go func() {
				_, ok := <-signal
				assert.False(t, ok)
				i++
				assert.Equal(t, 2, i)
			}()
			assert.Equal(t, 0, i)

			go func() {
				after := time.After(1 * time.Second)
				<-after

				i++
				assert.Equal(t, 1, i)
				executed = true
				in.h.End()
			}()
		}).
		Serve()
	assert.True(t, executed)
}

func TestPreparer_Run(t *testing.T) {
	gone.Prepare().Run(func(in struct {
		h gone.Heaven `gone:"*"`
	}) {
		assert.NotNil(t, in.h)
	})
}

func TestPreparer_Serve1(t *testing.T) {
	gone.AfterStopSignalWaitSecond = 0
	gone.Prepare().Serve(func(in struct {
		h gone.Heaven `gone:"*"`
	}) {
		assert.NotNil(t, in.h)
		go func() {
			time.Sleep(10 * time.Millisecond)
			in.h.End()
		}()
	})
}

func TestHooksOrder(t *testing.T) {
	var orders []int
	var funcs []func()

	for i := 0; i < 3; i++ {
		func(i int) {
			funcs = append(funcs, func() {
				orders = append(orders, i)
			})
		}(i)
	}

	t.Run("BeforeStart", func(t *testing.T) {
		orders = nil
		prepare := gone.Prepare()

		for _, fn := range funcs {
			prepare.BeforeStart(fn)
		}
		prepare.Run()
		assert.Equal(t, []int{2, 1, 0}, orders)
	})

	t.Run("AfterStart", func(t *testing.T) {
		orders = nil
		prepare := gone.Prepare()

		for _, fn := range funcs {
			prepare.AfterStart(fn)
		}
		prepare.Run()
		assert.Equal(t, []int{0, 1, 2}, orders)
	})

	t.Run("BeforeStop", func(t *testing.T) {
		orders = nil
		prepare := gone.Prepare()

		for _, fn := range funcs {
			prepare.BeforeStop(fn)
		}
		prepare.Run()
		assert.Equal(t, []int{2, 1, 0}, orders)
	})

	t.Run("AfterStop", func(t *testing.T) {
		orders = nil
		prepare := gone.Prepare()

		for _, fn := range funcs {
			prepare.AfterStop(fn)
		}
		prepare.Run()
		assert.Equal(t, []int{0, 1, 2}, orders)
	})
}

type SayName interface {
	SayMyName() string
}

type TestGoner struct {
	gone.Flag
	Name string
}

func (g *TestGoner) SayMyName() string {
	return g.Name
}

type TestNamedGoner struct {
	gone.Flag
	Name string
}

func (g *TestNamedGoner) GetGonerId() gone.GonerId {
	return "test-named-goner"
}

func (g *TestNamedGoner) SayMyName() string {
	return g.Name
}

func TestPreparer_Load(t *testing.T) {
	t1 := TestGoner{Name: "test"}
	t2 := TestNamedGoner{Name: "test-named"}
	gone.
		Default.
		Load(&t1).
		Bury(&t2).
		LoadPriest(func(cemetery gone.Cemetery) error {
			cemetery.Bury(&t1, gone.GonerId("t1"))
			return nil
		}).
		Run(func(in struct {
			t1 *TestGoner `gone:"*"`
			t2 SayName    `gone:"test-named-goner"`
			tx SayName    `gone:"t1"`
		}) {
			assert.Equal(t, in.t1, &t1)
			assert.Equal(t, in.t2, &t2)
			assert.Equal(t, in.tx, &t1)
		})
}
