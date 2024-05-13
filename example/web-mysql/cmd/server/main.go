package main

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
	"web-mysql/internal"
)

func main() {
	//给启动流程增加traceId
	tracer.SetTraceId("", func() {
		gone.Serve(internal.MasterPriest)
	})
}
