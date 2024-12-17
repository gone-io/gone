package tracer

import (
	"github.com/gone-io/gone"
	"github.com/google/uuid"
	"github.com/jtolds/gls"
	"sync"
)

var load = gone.OnceLoad(func(loader gone.Loader) error {
	return loader.Load(
		&tracer{},
		gone.IsDefault(new(gone.Tracer)),
		gone.LazyFill(),
	)
})

func Load(loader gone.Loader) error {
	return load(loader)
}

// Priest Deprecated, use Load instead
func Priest(loader gone.Loader) error {
	return Load(loader)
}

type tracer struct {
	gone.Flag
	gone.Logger `gone:"*"`
}

var xMap sync.Map

func (t *tracer) Name() string {
	return "tracer"
}

func (t *tracer) SetTraceId(traceId string, cb func()) {
	SetTraceId(traceId, cb, t.Warnf)
}

func (t *tracer) GetTraceId() (traceId string) {
	return GetTraceId()
}

func (t *tracer) Go(cb func()) {
	traceId := t.GetTraceId()
	if traceId == "" {
		go cb()
	} else {
		go func() {
			t.SetTraceId(traceId, cb)
		}()
	}
}

func (t *tracer) Recover() {
	if err := recover(); err != nil {
		t.Errorf("handle panic: %v, %s", err, gone.PanicTrace(2, 1))
	}
}

func (t *tracer) RecoverSetTraceId(traceId string, fn func()) {
	t.SetTraceId(traceId, func() {
		defer t.Recover()
		fn()
	})
}

func GetTraceId() (traceId string) {
	gls.EnsureGoroutineId(func(gid uint) {
		if v, ok := xMap.Load(gid); ok {
			traceId = v.(string)
		}
	})
	return
}

func SetTraceId(traceId string, cb func(), log ...func(format string, args ...any)) {
	id := GetTraceId()
	if "" != id {
		if len(log) > 0 {
			log[0]("SetTraceId not success for Having been set")
		}
		cb()
		return
	} else {
		if traceId == "" {
			traceId = uuid.New().String()
		}
		gls.EnsureGoroutineId(func(gid uint) {
			xMap.Store(gid, traceId)
			defer xMap.Delete(gid)
			cb()
		})
	}
}

//func GetGoroutineId() (gid uint64) {
//	var (
//		buf [64]byte
//		n   = runtime.Stack(buf[:], false)
//		stk = strings.TrimPrefix(string(buf[:n]), "goroutine ")
//	)
//	idField := strings.Fields(stk)[0]
//	var err error
//	gid, err = strconv.ParseUint(idField, 10, 64)
//	if err != nil {
//		panic(fmt.Errorf("can not get goroutine id: %v", err))
//	}
//	return
//}
