package tracer_test

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"gopkg.in/errgo.v2/errors"
	"testing"
)

func Test_tracer_Recover(t1 *testing.T) {
	gone.Prepare(tracer.Priest, config.Priest, logrus.Priest).AfterStart(func(in struct {
		tracer gone.Tracer `gone:"gone-tracer"`
	}) {
		in.tracer.RecoverSetTraceId("", func() {
			panic(errors.New("err"))
		})
	}).Run()
}
