package gin

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var xMap sync.Map

func NewHttInjector() (gone.Goner, gone.GonerId) {
	return &httpInjector{}, "http"
}

type httpInjector struct {
	gone.Flag
	tracer      tracer.Tracer `gone:"gone-tracer"`
	gone.Logger `gone:"gone-logger"`
}

func parseConfKeyValue(conf string) (key, value string) {
	before, after, found := strings.Cut(conf, "=")
	if found {
		return before, after
	} else {
		return before, ""
	}
}

func (s *httpInjector) SetContext(context *Context) (any, error) {
	return s.setContext(context, context.Next)
}

func (s *httpInjector) setContext(context *Context, next func()) (any, error) {
	traceId := s.tracer.GetTraceId()
	xMap.Store(traceId, context)
	defer xMap.Delete(traceId)
	next()
	return nil, nil
}

func (s *httpInjector) Suck(conf string, v reflect.Value, field reflect.StructField) error {
	traceId := s.tracer.GetTraceId()
	if traceId == "" {
		s.Errorf("inject field(%s) failed", field.Name)
		return NewInnerError("traceId is empty", http.StatusInternalServerError)
	}

	if x, ok := xMap.Load(traceId); ok {
		ctx := x.(*Context)
		kind, key := parseConfKeyValue(conf)
		if key == "" {
			key = field.Name
		}
		return s.inject(ctx, kind, key, v, field.Name)
	} else {
		s.Errorf("inject field(%s) failed", field.Name)
		return NewInnerError("cannot load context", http.StatusInternalServerError)
	}
}

const keyBody = "body"
const keyHeader = "header"
const keyParam = "param"
const keyQuery = "query"
const keyCookie = "cookie"

func unsupportedAttributeType(fieldName string) error {
	return NewInnerError(fmt.Sprintf("cannot inject %s，unsupported attribute type; ref doc: https://goner.fun/en/references/http-inject.md", fieldName), http.StatusInternalServerError)
}
func unsupportedKindConfigure(fieldName string) error {
	return NewInnerError(fmt.Sprintf("cannot inject %s，unsupported kind configure; ref doc: https://goner.fun/en/references/http-inject.md", fieldName), http.StatusInternalServerError)
}

func (s *httpInjector) inject(ctx *Context, kind string, key string, v reflect.Value, fieldName string) error {
	switch kind {
	case "":
		return s.injectWithoutKind(ctx, v, fieldName)
	case keyBody:
		return s.injectBody(ctx, v, fieldName)
	default:
		return s.injectByKind(ctx, kind, key, v, fieldName)
	}
}

func (s *httpInjector) injectWithoutKind(ctx *Context, v reflect.Value, fieldName string) error {
	t := v.Type()
	switch t {
	case reflect.TypeOf(ctx):
		v.Set(reflect.ValueOf(ctx))

	case reflect.TypeOf(ctx).Elem():
		v.Set(reflect.ValueOf(ctx).Elem())

	case reflect.TypeOf(ctx.Request):
		v.Set(reflect.ValueOf(ctx.Request))

	case reflect.TypeOf(ctx.Request).Elem():
		v.Set(reflect.ValueOf(ctx.Request).Elem())

	case reflect.TypeOf(ctx.Request.URL):
		v.Set(reflect.ValueOf(ctx.Request.URL))

	case reflect.TypeOf(ctx.Request.URL).Elem():
		v.Set(reflect.ValueOf(ctx.Request.URL).Elem())

	case reflect.TypeOf(ctx.Request.Header):
		v.Set(reflect.ValueOf(ctx.Request.Header))
	default:
		if t.Kind() == reflect.Interface && reflect.TypeOf(ctx.Writer).Implements(t) {
			v.Set(reflect.ValueOf(ctx.Writer))
		} else {
			s.Errorf("inject field(%s) failed", fieldName)
			return unsupportedAttributeType(fieldName)
		}
	}
	return nil
}

func (s *httpInjector) injectBody(ctx *Context, v reflect.Value, fieldName string) error {
	t := v.Type()
	if !(t.Kind() == reflect.Struct || t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Struct) {
		s.Errorf("inject field(%s) failed", fieldName)
		return unsupportedAttributeType(fieldName)
	}

	if t.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		if err := ctx.ShouldBind(v.Interface()); err != nil {
			return NewParameterError(err.Error())
		}
	} else {
		body := reflect.New(t).Interface()
		if err := ctx.ShouldBind(body); err != nil {
			return NewParameterError(err.Error())
		}
		v.Set(reflect.ValueOf(body).Elem())
	}
	return nil
}

func (s *httpInjector) injectByKind(ctx *Context, kind, key string, v reflect.Value, fieldName string) error {
	switch kind {
	case keyHeader:
		value := ctx.Request.Header.Get(key)
		return s.parseStringValueAndInject(v, fieldName, value)
	case keyParam:
		value := ctx.Param(key)
		return s.parseStringValueAndInject(v, fieldName, value)
	case keyQuery:
		return s.injectQuery(ctx, v, fieldName, key)
	case keyCookie:
		value, err := ctx.Context.Cookie(key)
		if err != nil {
			return NewParameterError(err.Error())
		}
		return s.parseStringValueAndInject(v, fieldName, value)
	default:
		return unsupportedKindConfigure(fieldName)
	}
}

func (s *httpInjector) injectQuery(ctx *Context, v reflect.Value, fieldName string, key string) error {
	t := v.Type()
	switch t.Kind() {
	case reflect.Struct:
		body := reflect.New(t).Interface()
		if err := ctx.ShouldBindQuery(body); err != nil {
			return NewParameterError(err.Error())
		}
		v.Set(reflect.ValueOf(body).Elem())
		return nil
	case reflect.Pointer:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if err := ctx.ShouldBindQuery(v.Interface()); err != nil {
			return NewParameterError(err.Error())
		}
		return nil
	case reflect.Slice:
		return s.injectQueryArray(ctx, key, v, fieldName)
	default:
		value := ctx.Query(key)
		return s.parseStringValueAndInject(v, fieldName, value)
	}
}

func (s *httpInjector) injectQueryArray(ctx *Context, key string, v reflect.Value, fieldName string) error {
	values := ctx.QueryArray(key)

	kind := v.Type().Elem().Kind()

	switch kind {
	case reflect.String:
		v.Set(reflect.ValueOf(values))
	case reflect.Bool:
		for _, value := range values {
			v.Set(reflect.Append(v, reflect.ValueOf(value != "")))
		}
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		for _, value := range values {
			def, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return NewParameterError(err.Error())
			}

			switch kind {
			case reflect.Int:
				v.Set(reflect.Append(v, reflect.ValueOf(int(def))))
			case reflect.Int64:
				v.Set(reflect.Append(v, reflect.ValueOf(def)))
			case reflect.Int32:
				v.Set(reflect.Append(v, reflect.ValueOf(int32(def))))
			case reflect.Int16:
				v.Set(reflect.Append(v, reflect.ValueOf(int16(def))))
			case reflect.Int8:
				v.Set(reflect.Append(v, reflect.ValueOf(int8(def))))
			}
		}
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		for _, value := range values {
			def, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return NewParameterError(err.Error())
			}
			switch kind {
			case reflect.Uint:
				v.Set(reflect.Append(v, reflect.ValueOf(uint(def))))
			case reflect.Uint64:
				v.Set(reflect.Append(v, reflect.ValueOf(def)))
			case reflect.Uint32:
				v.Set(reflect.Append(v, reflect.ValueOf(uint32(def))))
			case reflect.Uint16:
				v.Set(reflect.Append(v, reflect.ValueOf(uint16(def))))
			case reflect.Uint8:
				v.Set(reflect.Append(v, reflect.ValueOf(uint8(def))))
			}
		}

	case reflect.Float64, reflect.Float32:
		for _, value := range values {
			def, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return NewParameterError(err.Error())
			}
			if kind == reflect.Float64 {
				v.Set(reflect.Append(v, reflect.ValueOf(def)))
			} else {
				v.Set(reflect.Append(v, reflect.ValueOf(float32(def))))
			}
		}
	default:
		return unsupportedAttributeType(fieldName)
	}
	return nil
}

func (s *httpInjector) parseStringValueAndInject(v reflect.Value, fieldName string, value string) error {
	t := v.Type()
	switch t.Kind() {
	case reflect.String:
		v.Set(reflect.ValueOf(value))
	case reflect.Bool:
		v.Set(reflect.ValueOf(value != "" && value != "0" && value != "false"))
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		def, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return NewParameterError(err.Error())
		}
		v.SetInt(def)
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		def, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return NewParameterError(err.Error())
		}
		v.SetUint(def)

	case reflect.Float64, reflect.Float32:
		def, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return NewParameterError(err.Error())
		}
		v.SetFloat(def)
	default:
		s.Errorf("inject field(%s) failed", fieldName)
		return unsupportedAttributeType(fieldName)
	}
	return nil
}
