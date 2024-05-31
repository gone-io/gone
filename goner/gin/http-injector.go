package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

func NewHttInjector() (gone.Goner, gone.GonerId, gone.GonerOption) {
	return &httpInjector{}, gone.IdHttpInjector, gone.IsDefault(true)
}

type BindFunc func(context *gin.Context) error

type httpInjector struct {
	gone.Flag
	tracer      tracer.Tracer `gone:"gone-tracer"`
	gone.Logger `gone:"gone-logger"`

	bindFuncs      []BindFunc
	isInjectedBody bool
}

func parseConfKeyValue(conf string) (key, value string) {
	before, after, found := strings.Cut(conf, "=")
	if found {
		return before, after
	} else {
		return before, ""
	}
}

func (s *httpInjector) StartCollectBindFuncs() {
	s.bindFuncs = make([]BindFunc, 0)
	s.isInjectedBody = false
}

func (s *httpInjector) CollectBindFuncs() []BindFunc {
	return s.bindFuncs
}

func (s *httpInjector) Suck(conf string, v reflect.Value, field reflect.StructField) error {
	kind, key := parseConfKeyValue(conf)
	if key == "" {
		key = field.Name
	}
	fn, err := s.inject(kind, key, v, field.Name)
	if err != nil {
		return err
	}
	s.bindFuncs = append(s.bindFuncs, fn)
	return nil
}

const keyBody = "body"
const keyHeader = "header"
const keyParam = "param"
const keyQuery = "query"
const keyCookie = "cookie"

func unsupportedAttributeType(fieldName string) error {
	return NewInnerError(fmt.Sprintf("cannot inject %s，unsupported attribute type; ref doc: https://goner.fun/en/references/http-inject.md", fieldName), gone.InjectError)
}
func unsupportedKindConfigure(fieldName string) error {
	return NewInnerError(fmt.Sprintf("cannot inject %s，unsupported kind configure; ref doc: https://goner.fun/en/references/http-inject.md", fieldName), gone.InjectError)
}

func cannotInjectBodyMoreThanOnce(fieldName string) error {
	return NewInnerError(fmt.Sprintf("cannot inject %s，http body inject only support inject once; ref doc: https://goner.fun/en/references/http-inject.md", fieldName), gone.InjectError)
}

func (s *httpInjector) inject(kind string, key string, v reflect.Value, fieldName string) (fn BindFunc, err error) {
	if kind == "" {
		return s.injectWithoutKind(v, fieldName)
	}
	return s.injectByKind(kind, key, v, fieldName)
}

var ctxPtr *gin.Context
var ctxPointType = reflect.TypeOf(ctxPtr)
var ctxType = ctxPointType.Elem()

var requestPtr *http.Request
var requestType = reflect.TypeOf(requestPtr)
var requestPointType = requestType.Elem()

var urlPtr *url.URL
var urlType = reflect.TypeOf(urlPtr)
var urlPointType = urlType.Elem()

var header http.Header
var headerType = reflect.TypeOf(header)

var writerPtr gin.ResponseWriter
var writerType = reflect.TypeOf(writerPtr)

func (s *httpInjector) injectWithoutKind(v reflect.Value, fieldName string) (fn BindFunc, err error) {
	t := v.Type()
	switch t {
	case ctxPointType:
		return func(ctx *gin.Context) error {
			v.Set(reflect.ValueOf(ctx))
			return nil
		}, nil

	case ctxType:
		return func(ctx *gin.Context) error {
			v.Set(reflect.ValueOf(ctx).Elem())
			return nil
		}, nil

	case requestType:
		return func(ctx *gin.Context) error {
			v.Set(reflect.ValueOf(ctx.Request))
			return nil
		}, nil

	case requestPointType:
		return func(ctx *gin.Context) error {
			v.Set(reflect.ValueOf(ctx.Request).Elem())
			return nil
		}, nil

	case urlType:
		return func(ctx *gin.Context) error {
			v.Set(reflect.ValueOf(ctx.Request.URL))
			return nil
		}, nil

	case urlPointType:
		return func(ctx *gin.Context) error {
			v.Set(reflect.ValueOf(ctx.Request.URL).Elem())
			return nil
		}, nil

	case headerType:
		return func(ctx *gin.Context) error {
			v.Set(reflect.ValueOf(ctx.Request.Header))
			return nil
		}, nil

	default:
		if t.Kind() == reflect.Interface && writerType.Implements(t) {
			return func(ctx *gin.Context) error {
				v.Set(reflect.ValueOf(ctx.Writer))
				return nil
			}, nil
		} else {
			s.Errorf("inject field(%s) failed", fieldName)
			return nil, unsupportedAttributeType(fieldName)
		}
	}
}

func (s *httpInjector) injectBody(v reflect.Value, fieldName string) (fn BindFunc, err error) {
	if s.isInjectedBody {
		return nil, cannotInjectBodyMoreThanOnce(fieldName)
	}

	t := v.Type()
	if !(t.Kind() == reflect.Struct || t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Struct) {
		s.Errorf("inject field(%s) failed", fieldName)
		return nil, unsupportedAttributeType(fieldName)
	}

	if t.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		return func(ctx *gin.Context) error {
			if err := ctx.ShouldBind(v.Interface()); err != nil {
				return NewParameterError(err.Error())
			}
			return nil
		}, nil
	} else {
		return func(ctx *gin.Context) error {
			body := reflect.New(t).Interface()

			if err := ctx.ShouldBind(body); err != nil {
				return NewParameterError(err.Error())
			}
			v.Set(reflect.ValueOf(body).Elem())
			return nil
		}, nil
	}
}

func (s *httpInjector) injectByKind(kind, key string, v reflect.Value, fieldName string) (fn BindFunc, err error) {
	switch kind {
	case keyHeader, keyParam, keyCookie:
		return s.parseStringValueAndInject(v, fieldName, kind, key)
	case keyQuery:
		return s.injectQuery(v, fieldName, key)
	case keyBody:
		return s.injectBody(v, fieldName)
	default:
		return nil, unsupportedKindConfigure(fieldName)
	}
}

func (s *httpInjector) injectQuery(v reflect.Value, fieldName string, key string) (fn BindFunc, err error) {
	t := v.Type()
	switch t.Kind() {
	case reflect.Struct:
		return func(ctx *gin.Context) error {
			body := reflect.New(t).Interface()
			if err := ctx.ShouldBindQuery(body); err != nil {
				return NewParameterError(err.Error())
			}
			v.Set(reflect.ValueOf(body).Elem())
			return nil
		}, nil

	case reflect.Pointer:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		return func(ctx *gin.Context) error {
			if err := ctx.ShouldBindQuery(v.Interface()); err != nil {
				return NewParameterError(err.Error())
			}
			return nil
		}, nil
	case reflect.Slice:
		return s.injectQueryArray(key, v, fieldName)
	default:
		return s.parseStringValueAndInject(v, fieldName, keyQuery, key)
	}
}

func (s *httpInjector) injectQueryArray(key string, v reflect.Value, fieldName string) (fn BindFunc, err error) {
	kind := v.Type().Elem().Kind()

	switch kind {
	case reflect.String:
		return func(ctx *gin.Context) error {
			values := ctx.QueryArray(key)
			v.Set(reflect.ValueOf(values))
			return nil
		}, nil

	case reflect.Bool:
		return func(ctx *gin.Context) error {
			values := ctx.QueryArray(key)
			for _, value := range values {
				v.Set(reflect.Append(v, reflect.ValueOf(value != "")))
			}
			return nil
		}, nil

	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return func(ctx *gin.Context) error {
			values := ctx.QueryArray(key)
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
			return nil
		}, nil

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return func(ctx *gin.Context) error {
			values := ctx.QueryArray(key)
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
			return nil
		}, nil

	case reflect.Float64, reflect.Float32:
		return func(ctx *gin.Context) error {
			values := ctx.QueryArray(key)
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
			return nil
		}, nil
	default:
		return nil, unsupportedAttributeType(fieldName)
	}
}

func (s *httpInjector) parseStringValueAndInject(v reflect.Value, fieldName string, kind string, key string) (fn BindFunc, err error) {
	var parser func(context *gin.Context) (string, error)

	switch kind {
	case keyHeader:
		parser = func(context *gin.Context) (string, error) {
			return context.GetHeader(key), nil
		}
	case keyParam:
		parser = func(context *gin.Context) (string, error) {
			return context.Param(key), nil
		}
	case keyCookie:
		parser = func(context *gin.Context) (string, error) {
			return context.Cookie(key)
		}
	case keyQuery:
		parser = func(context *gin.Context) (string, error) {
			return context.Query(key), nil
		}
	default:
		return nil, unsupportedKindConfigure(fieldName)
	}

	t := v.Type()
	switch t.Kind() {
	case reflect.String:
		return func(context *gin.Context) error {
			value, err := parser(context)
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(value))
			return nil
		}, nil

	case reflect.Bool:
		return func(context *gin.Context) error {
			value, err := parser(context)
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(value != "" && value != "0" && value != "false"))
			return nil
		}, nil
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return func(context *gin.Context) error {
			value, err := parser(context)
			if err != nil {
				return err
			}

			def, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return NewParameterError(err.Error())
			}
			v.SetInt(def)

			return nil
		}, nil

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return func(context *gin.Context) error {
			value, err := parser(context)
			if err != nil {
				return err
			}

			def, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return NewParameterError(err.Error())
			}
			v.SetUint(def)

			return nil
		}, nil

	case reflect.Float64, reflect.Float32:
		return func(context *gin.Context) error {
			value, err := parser(context)
			if err != nil {
				return err
			}

			def, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return NewParameterError(err.Error())
			}
			v.SetFloat(def)

			return nil
		}, nil

	default:
		return nil, unsupportedAttributeType(fieldName)
	}
}
