package main

import (
	"demo-structure/internal"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
)

func main() {
	tracer.SetTraceId("", func() {
		gone.Serve(internal.MasterPriest)
	})
}
