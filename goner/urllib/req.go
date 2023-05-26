package urllib

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/imroc/req/v3"
)

const TraceIdHeaderKey = "X-Trace-ID"

func NewReq() (gone.Goner, gone.GonerId) {
	return &r{}, gone.IdGoneReq
}

type r struct {
	gone.Flag
	*req.Client
	tracer.Tracer `gone:"gone-tracer"`
}

func (r *r) AfterRevive() gone.AfterReviveError {
	r.Client = req.C()

	r.Client.WrapRoundTripFunc(func(rt req.RoundTripper) req.RoundTripFunc {
		return func(req *req.Request) (resp *req.Response, err error) {
			tracerId := r.GetTraceId()
			//传递traceId
			req.SetHeader(TraceIdHeaderKey, tracerId)
			resp, err = rt.RoundTrip(req)
			return
		}
	})

	return nil
}
