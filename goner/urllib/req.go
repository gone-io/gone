package urllib

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/imroc/req/v3"
)

const TraceIdHeaderKey = "X-Trace-ID"

func NewReq() (gone.Goner, gone.GonerId, gone.GonerOption) {
	return &r{}, gone.IdGoneReq, gone.IsDefault(new(Client))
}

type r struct {
	gone.Flag
	*req.Client
	tracer.Tracer `gone:"gone-tracer"`
}

func (r *r) trip(rt req.RoundTripper) req.RoundTripFunc {
	return func(req *req.Request) (resp *req.Response, err error) {
		tracerId := r.GetTraceId()
		//传递traceId
		req.SetHeader(TraceIdHeaderKey, tracerId)
		resp, err = rt.RoundTrip(req)
		return
	}
}

func (r *r) AfterRevive() gone.AfterReviveError {
	r.Client = req.C()

	r.Client.WrapRoundTripFunc(r.trip)

	return nil
}

func (r *r) C() *req.Client {
	c := req.C()
	c.WrapRoundTripFunc(r.trip)
	return c
}
