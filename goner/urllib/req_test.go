package urllib

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_r_AfterRevive(t *testing.T) {
	gone.Prepare(tracer.Priest, logrus.Priest, config.Priest).AfterStart(func(in struct {
		tracer tracer.Tracer `gone:"gone-tracer"`
	}) {
		g := r{
			Tracer: in.tracer,
		}
		err := g.AfterRevive()
		assert.Nil(t, err)
	}).Run()
}

func Test_r_trip(t *testing.T) {
	gone.Prepare(tracer.Priest, logrus.Priest, config.Priest).AfterStart(func(in struct {
		tracer tracer.Tracer `gone:"gone-tracer"`
	}) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		tripper := NewMockRoundTripper(controller)

		in.tracer.SetTraceId("xxxx", func() {
			tripper.EXPECT().RoundTrip(gomock.Any()).Do(func(req *req.Request) {
				traceId := req.Headers.Get(TraceIdHeaderKey)
				assert.Equal(t, "xxxx", traceId)
			}).Return(nil, nil)

			g := r{
				Tracer: in.tracer,
			}
			trip := g.trip(tripper)
			_, err := trip(&req.Request{})
			assert.Nil(t, err)
		})
	}).Run()
}
