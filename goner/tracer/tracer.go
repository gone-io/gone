package tracer

import (
	"github.com/gone-io/gone"
	"github.com/google/uuid"
	"github.com/jtolds/gls"
)

type tracer struct {
	gone.Flag
	mgr         *gls.ContextManager
	traceIdKey  gls.ContextKey
	gone.Logger `gone:"gone-logger"`
}

func (t *tracer) SetTraceId(traceId string, fn func()) {
	if traceId == "" {
		traceId = uuid.New().String()
	}
	t.mgr.SetValues(gls.Values{t.traceIdKey: traceId}, fn)
}
func (t *tracer) GetTraceId() string {
	if traceId, ok := t.mgr.GetValue(t.traceIdKey); ok {
		return traceId.(string)
	} else {
		return ""
	}
}

func (t *tracer) Go(fn func()) {
	gls.Go(func() {
		defer t.Recover()
		if "" == t.GetTraceId() {
			t.SetTraceId("", fn)
		} else {
			fn()
		}
	})
}

func (t *tracer) Recover() {
	if err := recover(); err != nil {
		t.Errorf("handle panic: %v, %s", err, gone.PanicTrace(2))
	}
}

func (t *tracer) RecoverSetTraceId(traceId string, fn func()) {
	t.SetTraceId(traceId, func() {
		t.Recover()
		fn()
	})
}
