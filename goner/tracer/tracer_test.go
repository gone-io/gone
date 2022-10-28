package tracer

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTracer(t *testing.T) {
	gone.Test(func(tracer *tracer) {
		traceId := "test-id"
		i := 100
		var fn func()
		fn = func() {
			i--
			if i == 0 {
				return
			}

			id := tracer.GetTraceId()
			assert.Equal(t, id, traceId)
			tracer.Go(func() {
				id := tracer.GetTraceId()
				assert.Equal(t, id, traceId)
			})
		}

		tracer.SetTraceId(traceId, fn)
	}, Priest, func(cemetery gone.Cemetery) error {
		cemetery.Bury(gone.NewSimpleLogger())
		return nil
	})
}

func BenchmarkTracer_GetTraceId(b *testing.B) {
	gone.Test(func(tracer *tracer) {
		traceId := "test-id"
		i := 20 //gls.EnsureGoroutineId 是通过查询堆栈标记完成的，堆栈的深度影响库的性能
		var fn func() int
		fn = func() int {
			i--
			if i == 0 {
				for i := 0; i < b.N; i++ {
					id := tracer.GetTraceId()
					assert.Equal(b, id, traceId)
				}
				return i
			}
			//golang貌似不会做尾递归优化
			return fn()
		}

		tracer.SetTraceId(traceId, func() {
			fn()
		})
	}, Priest, func(cemetery gone.Cemetery) error {
		cemetery.Bury(gone.NewSimpleLogger())
		return nil
	})
}
