package gin

import (
	"errors"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

var xMap sync.Map

func NewHttInjector() (gone.Goner, gone.GonerId) {
	return &httpInjector{}, "http"
}

type httpInjector struct {
	gone.Flag
	tracer        tracer.Tracer `gone:"gone-tracer"`
	logrus.Logger `gone:"gone-logger"`
}

func (s *httpInjector) Suck(conf string, v reflect.Value) gone.SuckError {
	traceId := s.tracer.GetTraceId()
	if traceId == "" {
		return NewInnerError("traceId is empty", http.StatusInternalServerError)
	}

	if x, ok := xMap.Load(traceId); ok {
		ctx := x.(*Context)

		s.Info("conf:%s", conf)
		s.Info("v:%v", v)
		s.Info("ctx:%v", ctx)
		//todo
		split := strings.Split(conf, "=")

		if len(split) < 1 || len(split) > 2 {
			return errors.New("gone-http tag is error")
		}
		var key, value string
		key = split[0]
		if len(split) == 2 {
			value = split[1]
		}

		switch key {
		case "query":
			if value == "" {
				//v.Set(reflect.ValueOf(ctx.Request.URL.Query()))
				//todo
			} else {
				v.Set(reflect.ValueOf(ctx.Query(value)))
			}
		case "cookie":
			//todo

		case "header":
			//todo

		case "auth":
			//todo

		case "form":
			//todo

		case "host":
			v.Set(reflect.ValueOf(ctx.Request.Host))
		case "url":
			v.Set(reflect.ValueOf(ctx.Request.URL.String()))
		case "path":
			v.Set(reflect.ValueOf(ctx.Request.URL.Path))

		case "body":
			//todo
			//err := ctx.ShouldBindJSON(v.Interface())
			//if err != nil {
			//	return NewParameterError(err.Error())
			//}

		case "context":
			//todo
			//v.Set(reflect.ValueOf(ctx))

		case "headers":
			v.Set(reflect.ValueOf(ctx.Request.Header))
		}

		return nil
	} else {
		return NewInnerError("cannot load context", http.StatusInternalServerError)
	}
}
