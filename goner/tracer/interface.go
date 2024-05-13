package tracer

import "github.com/gone-io/gone"

// Tracer 日志追踪，用于给同一个调用链路赋予统一的traceId，方便日志追踪
// Deprecated use gone.Tracer instead
type Tracer = gone.Tracer

//type Tracer interface {
//
//	//SetTraceId 设置TraceId；给调用函数设置一个`traceId`,如果traceId为空字符串，将生成自动一个
//	SetTraceId(traceId string, fn func())
//
//	//GetTraceId 获取当前协程的traceId
//	GetTraceId() string
//
//	//Go 开启一个新的协程，用于替代语法操作`go fn()`,在新的协程中将自动携带当前协程的`traceId`
//	Go(fn func())
//
//	//Recover 捕获 panic
//	Recover()
//
//	//RecoverSetTraceId SetTraceId 且 捕获 panic
//	RecoverSetTraceId(traceId string, fn func())
//}
