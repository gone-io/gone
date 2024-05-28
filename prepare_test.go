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
