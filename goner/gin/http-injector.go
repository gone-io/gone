package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

func NewHttInjector() (gone.Goner, gone.GonerId, gone.GonerOption) {
	return &httpInjector{
		bindFuncs: make([]BindFieldFunc, 0),
	}, gone.IdHttpInjector, gone.IsDefault(true)
}

type httpInjector struct {
	gone.Flag

	bindFuncs      []BindFieldFunc
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

func (s *httpInjector) StartBindFuncs() {
	s.bindFuncs = make([]BindFieldFunc, 0)
	s.isInjectedBody = false
}

func (s *httpInjector) CollectBindFuncs() []BindFieldFunc {
	return s.bindFuncs
}

func (s *httpInjector) BindFuncs() BindStructFunc {
	funcs := s.CollectBindFuncs()
	return func(context *gin.Context, arg any, T reflect.Type) (reflect.Value, error) {
		v := reflect.ValueOf(&arg).Elem()
		v = reflect.NewAt(T, unsafe.Pointer(v.UnsafeAddr())).Elem()

		for _, fn := range funcs {
			err := fn(context, v)
			if err != nil {
				return v, err
			}
		}
		return v, nil
	}
}

func (s *httpInjector) Suck(conf string, v reflect.Value, field reflect.StructField) error {
	kind, key := parseConfKeyValue(conf)
	if key == "" {
		key = field.Name
	}
	fn, err := s.inject(kind, key, field)
	if err != nil {
		return err
	}

	//index := field.Index()
	s.bindFuncs = append(s.bindFuncs, fn)
	return nil
}

const keyBody = "body"
const keyHeader = "header"
const keyParam = "param"
const keyQuery = "query"
const keyCookie = "cookie"

func unsupportedAttributeType(fieldName string) error {
	return NewInnerError(fmt.Sprintf("cannot inject %s，unsupported attribute type; ref doc: https://goner.fun/references/http-inject.html", fieldName), gone.InjectError)
}
func unsupportedKindConfigure(fieldName string) error {
	return NewInnerError(fmt.Sprintf("cannot inject %s，unsupported kind configure; ref doc: https://goner.fun/references/http-inject.html", fieldName), gone.InjectError)
}

func injectParseStringParameterError(k reflect.Kind, kind, key string, err error) gone.Error {
	return NewParameterError(
		fmt.Sprintf("%s parameter %s required %s;parse error: %s", kind, key, k.String(), err.Error()),
	)
}

func cannotInjectBodyMoreThanOnce(fieldName string) error {
	return NewInnerError(fmt.Sprintf("cannot inject %s，http body inject only support inject once; ref doc: https://goner.fun/en/references/http-inject.md", fieldName), gone.InjectError)
}

func (s *httpInjector) inject(kind string, key string, field reflect.StructField) (fn BindFieldFunc, err error) {
	if kind == "" {
		return s.injectByType(field)
	}
	return s.injectByKind(kind, key, field)
}

//var ctxPtr *gin.Context
//var ctxPointType = reflect.TypeOf(ctxPtr)
//var ctxType = ctxPointType.Elem()

var requestPtr *http.Request
var requestType = reflect.TypeOf(requestPtr)
var requestPointType = requestType.Elem()

var urlPtr *url.URL
var urlType = reflect.TypeOf(urlPtr)
var urlPointType = urlType.Elem()

var header http.Header
var headerType = reflect.TypeOf(header)

var writerPtr *gin.ResponseWriter
var writerType = reflect.TypeOf(writerPtr).Elem()

func (s *httpInjector) injectByType(field reflect.StructField) (fn BindFieldFunc, err error) {
	t := field.Type
	switch t {
	case ctxPointType:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			v.Set(reflect.ValueOf(ctx))
			return nil
		}, nil

	case ctxType:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			v.Set(reflect.ValueOf(ctx).Elem())
			return nil
		}, nil

	case goneContextPointType:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			v.Set(reflect.ValueOf(&gone.Context{Context: ctx}))
			return nil
		}, nil

	case goneContextType:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			v.Set(reflect.ValueOf(gone.Context{Context: ctx}))
			return nil
		}, nil

	case requestType:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			v.Set(reflect.ValueOf(ctx.Request))
			return nil
		}, nil

	case requestPointType:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			v.Set(reflect.ValueOf(ctx.Request).Elem())
			return nil
		}, nil

	case urlType:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			v.Set(reflect.ValueOf(ctx.Request.URL))
			return nil
		}, nil

	case urlPointType:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			v.Set(reflect.ValueOf(ctx.Request.URL).Elem())
			return nil
		}, nil

	case headerType:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			v.Set(reflect.ValueOf(ctx.Request.Header))
			return nil
		}, nil

	case writerType:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			v.Set(reflect.ValueOf(ctx.Writer))
			return nil
		}, nil
	default:
		return nil, unsupportedAttributeType(field.Name)
	}
}

func (s *httpInjector) injectBody(kind, key string, field reflect.StructField) (fn BindFieldFunc, err error) {
	if s.isInjectedBody {
		return nil, cannotInjectBodyMoreThanOnce(field.Name)
	}

	t := field.Type
	switch t.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			body := reflect.New(t).Interface()

			if err := ctx.ShouldBind(body); err != nil {
				return NewParameterError(err.Error())
			}
			v.Set(reflect.ValueOf(body).Elem())
			return nil
		}, nil
	case reflect.Pointer:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			if err := ctx.ShouldBind(v.Interface()); err != nil {
				return NewParameterError(err.Error())
			}
			return nil
		}, nil
	default:
		return nil, unsupportedAttributeType(field.Name)
	}
}

func (s *httpInjector) injectByKind(kind, key string, field reflect.StructField) (fn BindFieldFunc, err error) {
	switch kind {
	case keyHeader, keyParam, keyCookie:
		return s.parseStringValueAndInject(kind, key, field)
	case keyQuery:
		return s.injectQuery(kind, key, field)
	case keyBody:
		return s.injectBody(kind, key, field)
	default:
		return nil, unsupportedKindConfigure(field.Name)
	}
}

var queryMapType = reflect.TypeOf(map[string]string{})

func (s *httpInjector) injectQuery(kind, key string, field reflect.StructField) (fn BindFieldFunc, err error) {
	t := field.Type
	switch t.Kind() {
	case reflect.Struct:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			body := reflect.New(t).Interface()
			if err := ctx.ShouldBindQuery(body); err != nil {
				return NewParameterError(err.Error())
			}
			v.Set(reflect.ValueOf(body).Elem())
			return nil
		}, nil

	case reflect.Pointer:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			if err := ctx.ShouldBindQuery(v.Interface()); err != nil {
				return NewParameterError(err.Error())
			}
			return nil
		}, nil

	case reflect.Map:
		if t == queryMapType {
			return func(ctx *gin.Context, structVale reflect.Value) error {
				v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
				dict := ctx.QueryMap(key)
				v.Set(reflect.ValueOf(dict))
				return nil
			}, nil
		}
		return nil, unsupportedAttributeType(field.Name)

	case reflect.Slice:
		return s.injectQueryArray(kind, key, field)

	default:
		return s.parseStringValueAndInject(kind, key, field)
	}
}

func (s *httpInjector) injectQueryArray(k, key string, field reflect.StructField) (fn BindFieldFunc, err error) {
	el := field.Type.Elem()

	kind := el.Kind()

	bits := bitSize(kind)
	switch kind {
	case reflect.String:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			values := ctx.QueryArray(key)
			v.Set(reflect.ValueOf(values))
			return nil
		}, nil

	case reflect.Bool:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			values := ctx.QueryArray(key)
			for _, value := range values {
				v.Set(reflect.Append(v, reflect.ValueOf(stringToBool(value))))
			}
			return nil
		}, nil

	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			values := ctx.QueryArray(key)
			for _, value := range values {
				def, err := strconv.ParseInt(value, 10, bits)
				if err != nil {
					return injectParseStringParameterError(kind, keyQuery, key, err)
				}
				v.Set(reflect.Append(v, reflect.ValueOf(def).Convert(el)))
			}
			return nil
		}, nil

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			values := ctx.QueryArray(key)
			for _, value := range values {
				def, err := strconv.ParseUint(value, 10, bits)
				if err != nil {
					return injectParseStringParameterError(kind, keyQuery, key, err)
				}
				v.Set(reflect.Append(v, reflect.ValueOf(def).Convert(el)))
			}
			return nil
		}, nil

	case reflect.Float64, reflect.Float32:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			values := ctx.QueryArray(key)
			for _, value := range values {
				def, err := strconv.ParseFloat(value, bits)
				if err != nil {
					return injectParseStringParameterError(kind, keyQuery, key, err)
				}
				v.Set(reflect.Append(v, reflect.ValueOf(def).Convert(el)))
			}
			return nil
		}, nil
	default:
		return nil, unsupportedAttributeType(field.Name)
	}
}

func (s *httpInjector) parseStringValueAndInject(kind, key string, field reflect.StructField) (fn BindFieldFunc, err error) {
	var parser func(context *gin.Context) (string, error)
	t := field.Type

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
			str, err := context.Cookie(key)
			if err != nil {
				return "", injectParseStringParameterError(t.Kind(), kind, key, err)
			}
			return str, nil
		}
	case keyQuery:
		parser = func(context *gin.Context) (string, error) {
			return context.Query(key), nil
		}
	default:
		return nil, unsupportedKindConfigure(field.Name)
	}

	bits := bitSize(t.Kind())

	switch t.Kind() {
	case reflect.String:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			value, err := parser(ctx)
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(value))
			return nil
		}, nil

	case reflect.Bool:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			value, err := parser(ctx)
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(stringToBool(value)))
			return nil
		}, nil
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			value, err := parser(ctx)
			if err != nil {
				return err
			}

			def, err := strconv.ParseInt(value, 10, bits)
			if err != nil {
				return injectParseStringParameterError(t.Kind(), kind, key, err)
			}
			v.SetInt(def)

			return nil
		}, nil

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			value, err := parser(ctx)
			if err != nil {
				return err
			}

			def, err := strconv.ParseUint(value, 10, bits)
			if err != nil {
				return injectParseStringParameterError(t.Kind(), kind, key, err)
			}
			v.SetUint(def)

			return nil
		}, nil

	case reflect.Float64, reflect.Float32:
		return func(ctx *gin.Context, structVale reflect.Value) error {
			v := fieldByIndexFromStructValue(structVale, field.Index, field.IsExported(), field.Type)
			value, err := parser(ctx)
			if err != nil {
				return err
			}

			def, err := strconv.ParseFloat(value, bits)
			if err != nil {
				return injectParseStringParameterError(t.Kind(), kind, key, err)
			}
			v.SetFloat(def)

			return nil
		}, nil

	default:
		return nil, unsupportedAttributeType(field.Name)
	}
}

func bitSize(kind reflect.Kind) int {
	switch kind {
	case reflect.Float64, reflect.Int64, reflect.Uint64:
		return 64
	case reflect.Float32, reflect.Int32, reflect.Uint32, reflect.Int, reflect.Uint:
		return 32
	case reflect.Int16, reflect.Uint16:
		return 16
	case reflect.Int8, reflect.Uint8:
		return 8
	default:
		return 32
	}
}

func stringToBool(value string) bool {
	return value != "" && value != "0" && value != "false"
}

func fieldByIndexFromStructValue(structValue reflect.Value, index []int, isExported bool, fieldType reflect.Type) reflect.Value {
	v := structValue.FieldByIndex(index)

	if !isExported {
		//黑魔法：让非导出字段可以访问
		v = reflect.NewAt(fieldType, unsafe.Pointer(v.UnsafeAddr())).Elem()
	}
	return v
}
