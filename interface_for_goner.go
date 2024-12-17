package gone

import (
	"github.com/soheilhy/cmux"
	"net"
	"xorm.io/xorm"
)

// CMuxServer cMux service，Used to multiplex the same port to listen for multiple protocols，ref：https://pkg.go.dev/github.com/soheilhy/cmux
type CMuxServer interface {
	Match(matcher ...cmux.Matcher) net.Listener
	MatchWithWriters(matcher ...cmux.MatchWriter) net.Listener
	GetAddress() string
}

// Tracer Log tracking, which is used to assign a unified traceId to the same call link to facilitate log tracking.
type Tracer interface {

	//SetTraceId to set `traceId` to the calling function. If traceId is an empty string, an automatic one will
	//be generated. TraceId can be obtained by using the GetTraceId () method in the calling function.
	SetTraceId(traceId string, fn func())

	//GetTraceId Get the traceId of the current goroutine
	GetTraceId() string

	//Go Start a new goroutine instead of `go func`, which can pass the traceid to the new goroutine.
	Go(fn func())

	//Recover use for catch panic in goroutine
	Recover()

	//RecoverSetTraceId SetTraceId and Recover
	RecoverSetTraceId(traceId string, fn func())
}

type XormEngine interface {
	xorm.EngineInterface
	Transaction(fn func(session xorm.Interface) error) error
	Sqlx(sql string, args ...any) *xorm.Session
	GetOriginEngine() xorm.EngineInterface
	SetPolicy(policy xorm.GroupPolicy)
}

const (
	RequestIdHeaderKey = "X-Request-Id"
	TraceIdHeaderKey   = "X-Trace-Id"
)
