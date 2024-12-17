package urllib

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/imroc/req/v3"
)

const TraceIdHeaderKey = gone.TraceIdHeaderKey

var load = gone.OnceLoad(func(loader gone.Loader) error {
	err := tracer.Load(loader)
	if err != nil {
		return gone.ToError(err)
	}
	return loader.Load(
		&r{},
		gone.IsDefault(new(Client)),
	)
})

func Load(loader gone.Loader) error {
	return load(loader)
}

// Priest Deprecated, use Load instead
func Priest(loader gone.Loader) error {
	return Load(loader)
}

type r struct {
	gone.Flag
	*req.Client
	gone.Tracer `gone:"*"`
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

func (r *r) Init() error {
	r.Client = req.C()
	r.Client.WrapRoundTripFunc(r.trip)
	return nil
}

func (r *r) C() *req.Client {
	c := req.C()
	c.WrapRoundTripFunc(r.trip)
	return c
}
