package gin

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/internal/json"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"
)

func Test_httpInjector_Suck(t *testing.T) {
	type Req struct {
		Str    string
		Number int
	}

	injector := httpInjector{}
	injector.StartBindFuncs()

	var req = &Req{}

	rv := reflect.ValueOf(req).Elem()
	for i := 0; i < rv.NumField(); i++ {
		err := injector.Suck("query", rv.Type().Field(i))
		assert.Nil(t, err)
	}
	funcs := injector.CollectBindFuncs()
	assert.Equal(t, len(funcs), 2)
}

func Test_httpInjector_inject(t *testing.T) {
	injector := httpInjector{}

	controller := gomock.NewController(t)
	defer controller.Finish()
	writer := NewMockResponseWriter(controller)
	writer.EXPECT().Written().AnyTimes()
	writer.EXPECT().WriteHeader(gomock.Any()).AnyTimes()
	writer.EXPECT().Header().Return(http.Header{}).AnyTimes()
	writer.EXPECT().Write(gomock.Any()).AnyTimes()

	const stringKey = "string-val"
	const stringVal = "gone is best"
	const NumberKey = "number"
	const NumberVal = 100

	const ErrorNumberKey = "error-number"
	const ErrorNumberVal = "1x00"

	Url, _ := url.Parse("https://goner.fun/zh/?page=1&pageSize=10&err=111x&map[user]=dapeng&map[age]=1024&arr=0&arr=1&arr=2&arr=3&" + stringKey + "=" + stringVal)

	context := gin.Context{
		Writer: writer,
		Request: &http.Request{
			URL: Url,
			Header: http.Header{
				"Content-Type": {"application/json"},
				"Accept":       {"application/json"},
				"Host":         {"goner.fun"},
				"Cookie":       {"key1=v1;key2=v2;" + stringKey + "=" + stringVal + ";"},
				stringKey:      {stringVal},
			},
		},
		Params: gin.Params{
			{
				Key:   stringKey,
				Value: stringVal,
			},
			{
				Key:   NumberKey,
				Value: strconv.Itoa(NumberVal),
			}, {
				Key:   ErrorNumberKey,
				Value: ErrorNumberVal,
			},
		},
	}

	type Struct struct {
		Page     int    `form:"page"`
		PageSize int    `form:"pageSize"`
		Arr      []int  `form:"arr"`
		Str      string `form:"string-val"`
	}

	type Body struct {
		X int    `json:"x,omitempty"`
		Y string `json:"y,omitempty"`
	}

	var req = &struct {
		Context    gin.Context
		ContextPtr *gin.Context

		GoneContext    gone.Context
		GoneContextPtr *gone.Context

		Request    http.Request
		RequestPtr *http.Request

		Url    url.URL
		UrlPtr *url.URL

		Header http.Header

		Writer gin.ResponseWriter

		ErrorType1 *gin.ResponseWriter

		Str   string
		Bool  bool
		Int   int
		Uint  uint
		Float float32

		Struct    Struct
		StructPrt *Struct

		QueryMap    map[string]string
		ErrQueryMap map[string]any

		StrSlice    []string
		BoolSlice   []bool
		IntSlice    []int
		UintSlice   []uint
		Uint16Slice []uint16
		FloatSlice  []float32

		Body       Body
		BodyPtr    *Body
		StringBody string
		BodyMap    map[string]any
		BodySlice  []any
	}{}

	tests := []struct {
		name      string
		fieldName string
		kind      string
		key       string

		wantFn  BindFieldFunc
		wantErr assert.ErrorAssertionFunc
		ctx     *gin.Context
		bindErr func(t assert.TestingT, err error)
		before  func()
	}{
		{
			name:      "inject gin.Context",
			fieldName: "Context",
			kind:      "",
			key:       "",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, context, req.Context)
			},
		},
		{
			name:      "inject *gin.Context",
			fieldName: "ContextPtr",
			kind:      "",
			key:       "",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, &context, req.ContextPtr)
			},
		},
		{
			name:      "inject gone.Context",
			fieldName: "GoneContext",
			kind:      "",
			key:       "",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, &context, req.GoneContext.Context)
			},
		},
		{
			name:      "inject *gone.Context",
			fieldName: "GoneContextPtr",
			kind:      "",
			key:       "",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, &context, req.GoneContextPtr.Context)
			},
		},
		{
			name:      "inject http.Request",
			fieldName: "Request",
			kind:      "",
			key:       "",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, *context.Request, req.Request)
			},
		},
		{
			name:      "inject *http.Request",
			fieldName: "RequestPtr",
			kind:      "",
			key:       "",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, context.Request, req.RequestPtr)
			},
		},
		{
			name:      "inject url.URL",
			fieldName: "Url",
			kind:      "",
			key:       "",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, *context.Request.URL, req.Url)
			},
		},
		{
			name:      "inject *url.URL",
			fieldName: "UrlPtr",
			kind:      "",
			key:       "",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, context.Request.URL, req.UrlPtr)
			},
		},
		{
			name:      "inject http.Header",
			fieldName: "Header",
			kind:      "",
			key:       "",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, context.Request.Header, req.Header)
			},
		},
		{
			name:      "inject gin.ResponseWriter",
			fieldName: "Writer",
			kind:      "",
			key:       "",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, context.Writer, req.Writer)
			},
		},
		{
			name:      "inject not support type",
			fieldName: "ErrorType1",
			kind:      "",
			key:       "",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Error(t, err)
				assert.Equal(t, err.(gone.Error).Msg(), unsupportedAttributeType("ErrorType1").(gone.Error).Msg())
				return false
			},

			bindErr: func(t assert.TestingT, err error) {},
		},

		{
			name:      "inject by kind, inject header string",
			fieldName: "Str",
			kind:      keyHeader,
			key:       stringKey,
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name:      "inject by kind, inject query string",
			fieldName: "Str",
			kind:      keyQuery,
			key:       stringKey,
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, req.Str, stringVal)
			},
		},
		{
			name:      "inject by kind, inject query *Struct",
			fieldName: "StructPrt",
			kind:      keyQuery,
			key:       stringKey,
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, *req.StructPrt, Struct{
					Page:     1,
					PageSize: 10,
					Arr:      []int{0, 1, 2, 3},
					Str:      stringVal,
				})
			},
		},
		{
			name:      "inject by kind, inject query Struct",
			fieldName: "Struct",
			kind:      keyQuery,
			key:       stringKey,
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, req.Struct, Struct{
					Page:     1,
					PageSize: 10,
					Arr:      []int{0, 1, 2, 3},
					Str:      stringVal,
				})
			},
		},

		{
			name:      "inject by kind, inject query map[string]string",
			fieldName: "QueryMap",
			kind:      keyQuery,
			key:       "map",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, map[string]string{
					"user": "dapeng",
					"age":  "1024",
				}, req.QueryMap)
			},
		},

		{
			name:      "inject by kind, inject query map[string]any error",
			fieldName: "ErrQueryMap",
			kind:      keyQuery,
			key:       "err",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Error(t, err)
				assert.Equal(t, gone.InjectError, err.(gone.Error).Code())
				return false
			},

			bindErr: func(t assert.TestingT, err error) {},
		},

		{
			name:      "inject by kind, inject query []string",
			fieldName: "StrSlice",
			kind:      keyQuery,
			key:       "arr",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, req.StrSlice, []string{"0", "1", "2", "3"})
			},
		},

		{
			name:      "inject by kind, inject query []bool",
			fieldName: "BoolSlice",
			kind:      keyQuery,
			key:       "arr",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []bool{false, true, false, false}, req.BoolSlice)
			},
		},
		{
			name:      "inject by kind, inject query []int",
			fieldName: "IntSlice",
			kind:      keyQuery,
			key:       "arr",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []int{0, 1, 2, 3}, req.IntSlice)
			},
		},
		{
			name:      "inject by kind, inject query []uint16",
			fieldName: "Uint16Slice",
			kind:      keyQuery,
			key:       "arr",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []uint16{0, 1, 2, 3}, req.Uint16Slice)
			},
		},
		{
			name:      "inject by kind, inject query []uint",
			fieldName: "UintSlice",
			kind:      keyQuery,
			key:       "arr",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []uint{0, 1, 2, 3}, req.UintSlice)
			},
		},
		{
			name:      "inject by kind, inject query []float32",
			fieldName: "FloatSlice",
			kind:      keyQuery,
			key:       "arr",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []float32{0, 1, 2, 3}, req.FloatSlice)
			},
		},

		{
			name:      "inject by kind, inject query []int error",
			fieldName: "IntSlice",
			kind:      keyQuery,
			key:       "err",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Error(t, err)
				assert.Equal(t, http.StatusBadRequest, err.(gone.Error).Code())
			},
		},
		{
			name:      "inject by kind, inject query []uint error",
			fieldName: "UintSlice",
			kind:      keyQuery,
			key:       "err",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Error(t, err)
				assert.Equal(t, http.StatusBadRequest, err.(gone.Error).Code())
			},
		},
		{
			name:      "inject by kind, inject query []float32 error",
			fieldName: "FloatSlice",
			kind:      keyQuery,
			key:       "err",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Error(t, err)
				assert.Equal(t, http.StatusBadRequest, err.(gone.Error).Code())
			},
		},

		{
			name:      "inject by kind, inject param string",
			fieldName: "Str",
			kind:      keyParam,
			key:       stringKey,
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, req.Str, stringVal)
			},
		},
		{
			name:      "inject by kind, inject cookie string",
			fieldName: "Str",
			kind:      keyCookie,
			key:       stringKey,
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, req.Str, stringVal)
			},
		},

		{
			name:      "inject by kind, inject cookie string error",
			fieldName: "Str",
			kind:      keyCookie,
			key:       "not existed",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Error(t, err)
				g, ok := err.(gone.Error)
				assert.True(t, ok)
				assert.Equal(t, g.Code(), http.StatusBadRequest)
			},
		},
		{
			name:      "inject by kind, inject cookie Bool error",
			fieldName: "Bool",
			kind:      keyCookie,
			key:       "not existed",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Error(t, err)
				g, ok := err.(gone.Error)
				assert.True(t, ok)
				assert.Equal(t, g.Code(), http.StatusBadRequest)
			},
		},

		{
			name:      "inject by kind, inject cookie Int error",
			fieldName: "Int",
			kind:      keyCookie,
			key:       "not existed",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Error(t, err)
				g, ok := err.(gone.Error)
				assert.True(t, ok)
				assert.Equal(t, g.Code(), http.StatusBadRequest)
			},
		},

		{
			name:      "inject by kind, inject cookie Uint error",
			fieldName: "Uint",
			kind:      keyCookie,
			key:       "not existed",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Error(t, err)
				g, ok := err.(gone.Error)
				assert.True(t, ok)
				assert.Equal(t, g.Code(), http.StatusBadRequest)
			},
		},

		{
			name:      "inject by kind, inject cookie Float error",
			fieldName: "Float",
			kind:      keyCookie,
			key:       "not existed",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Error(t, err)
				g, ok := err.(gone.Error)
				assert.True(t, ok)
				assert.Equal(t, g.Code(), http.StatusBadRequest)
			},
		},

		{
			name:      "inject by kind, unknown kind",
			fieldName: "Str",
			kind:      "x",
			key:       "not existed",
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Error(t, err)
				assert.Equal(t, err.(gone.Error).Msg(), unsupportedKindConfigure("Str", "x").(gone.Error).Msg())
				return false
			},

			bindErr: func(t assert.TestingT, err error) {},
		},

		{
			name:      "inject by kind, inject param bool",
			fieldName: "Bool",
			kind:      keyParam,
			key:       NumberKey,
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, req.Bool, false)
			},
		},

		{
			name:      "inject by kind, inject param Int",
			fieldName: "Int",
			kind:      keyParam,
			key:       NumberKey,
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, req.Int, NumberVal)
			},
		},
		{
			name:      "inject by kind, inject param Int Error",
			fieldName: "Int",
			kind:      keyParam,
			key:       ErrorNumberKey,
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Error(t, err)
				assert.Equal(t, err.(gone.Error).Code(), http.StatusBadRequest)
			},
		},

		{
			name:      "inject by kind, inject param Uint",
			fieldName: "Uint",
			kind:      keyParam,
			key:       NumberKey,
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, req.Uint, uint(NumberVal))
			},
		},
		{
			name:      "inject by kind, inject param Uint Error",
			fieldName: "Uint",
			kind:      keyParam,
			key:       ErrorNumberKey,
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Error(t, err)
				assert.Equal(t, err.(gone.Error).Code(), http.StatusBadRequest)
			},
		},

		{
			name:      "inject by kind, inject Float Uint",
			fieldName: "Float",
			kind:      keyParam,
			key:       NumberKey,
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, req.Float, float32(NumberVal))
			},
		},
		{
			name:      "inject by kind, inject param Float Error",
			fieldName: "Float",
			kind:      keyParam,
			key:       ErrorNumberKey,
			ctx:       &context,

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Error(t, err)
				assert.Equal(t, err.(gone.Error).Code(), http.StatusBadRequest)
			},
		},

		{
			name:      "inject by kind, inject body Struct",
			fieldName: "Body",
			kind:      keyBody,
			key:       stringKey,
			ctx:       &context,

			before: func() {
				body := Body{
					X: 100,
					Y: stringVal,
				}
				marshal, _ := json.Marshal(body)
				context.Request.Body = io.NopCloser(bytes.NewBuffer(marshal))
			},

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, Body{
					X: 100,
					Y: stringVal,
				}, req.Body)
			},
		},
		{
			name:      "inject by kind, inject body string",
			fieldName: "StringBody",
			kind:      keyBody,
			key:       stringKey,
			ctx:       &context,

			before: func() {
				body := Body{
					X: 100,
					Y: stringVal,
				}
				marshal, _ := json.Marshal(body)
				context.Request.Body = io.NopCloser(bytes.NewBuffer(marshal))
			},

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "{\"x\":100,\"y\":\"gone is best\"}", req.StringBody)
			},
		},
		{
			name:      "inject by kind, inject body Struct, parse error",
			fieldName: "Body",
			kind:      keyBody,
			key:       stringKey,
			ctx:       &context,

			before: func() {
				marshal := []byte(`{"X": 100i, "Y": 100}`)
				context.Request.Body = io.NopCloser(bytes.NewBuffer(marshal))
			},

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Error(t, err)
				assert.Equal(t, http.StatusBadRequest, err.(gone.Error).Code())
			},
		},
		{
			name:      "inject by kind, inject body Struct Pointer",
			fieldName: "BodyPtr",
			kind:      keyBody,
			key:       stringKey,
			ctx:       &context,

			before: func() {
				body := Body{
					X: 100,
					Y: stringVal,
				}
				marshal, _ := json.Marshal(body)
				context.Request.Body = io.NopCloser(bytes.NewBuffer(marshal))
			},

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, Body{
					X: 100,
					Y: stringVal,
				}, *req.BodyPtr)
			},
		},
		{
			name:      "inject by kind, inject body Struct Pointer, parse error",
			fieldName: "BodyPtr",
			kind:      keyBody,
			key:       stringKey,
			ctx:       &context,

			before: func() {
				marshal := []byte(`{"X": 100i, "Y": 100}`)
				context.Request.Body = io.NopCloser(bytes.NewBuffer(marshal))
			},

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Error(t, err)
				assert.Equal(t, http.StatusBadRequest, err.(gone.Error).Code())
			},
		},
		{
			name:      "inject by kind, inject body BodyMap",
			fieldName: "BodyMap",
			kind:      keyBody,
			key:       stringKey,
			ctx:       &context,

			before: func() {
				body := Body{
					X: 100,
					Y: stringVal,
				}
				marshal, _ := json.Marshal(body)
				context.Request.Body = io.NopCloser(bytes.NewBuffer(marshal))
			},

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, map[string]any{
					"x": float64(100),
					"y": stringVal,
				}, req.BodyMap)
			},
		},
		{
			name:      "inject by kind, inject body []any",
			fieldName: "BodySlice",
			kind:      keyBody,
			key:       stringKey,
			ctx:       &context,

			before: func() {
				body := []Body{{
					X: 100,
					Y: stringVal,
				}}
				marshal, _ := json.Marshal(body)
				context.Request.Body = io.NopCloser(bytes.NewBuffer(marshal))
			},

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []any{map[string]any{
					"x": float64(100),
					"y": stringVal,
				}}, req.BodySlice)
			},
		},
		{
			name:      "inject by kind, inject body []any test.yaml",
			fieldName: "BodySlice",
			kind:      keyBody,
			key:       stringKey,
			ctx:       &context,

			before: func() {
				body := []Body{{
					X: 100,
					Y: stringVal,
				}}
				marshal, _ := yaml.Marshal(body)
				context.Request.Body = io.NopCloser(bytes.NewBuffer(marshal))
				context.Request.Header = http.Header{
					"Content-Type": []string{"application/x-yaml"},
				}
			},

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},

			bindErr: func(t assert.TestingT, err error) {
				assert.Nil(t, err)
				assert.Equal(t, []any{map[string]any{
					"x": 100,
					"y": stringVal,
				}}, req.BodySlice)
			},
		},
		{
			name:      "inject by kind, inject body more than once",
			fieldName: "BodySlice",
			kind:      keyBody,
			key:       stringKey,
			ctx:       &context,

			before: func() {
				body := []Body{{
					X: 100,
					Y: stringVal,
				}}
				marshal, _ := yaml.Marshal(body)
				context.Request.Body = io.NopCloser(bytes.NewBuffer(marshal))
				context.Request.Header = http.Header{
					"Content-Type": []string{"application/x-test.yaml"},
				}
				injector.isInjectedBody = true
			},

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Error(t, err)
				return false
			},

			bindErr: func(t assert.TestingT, err error) {
			},
		},
		{
			name:      "inject by kind, inject body with unsupported type",
			fieldName: "Int",
			kind:      keyBody,
			key:       stringKey,
			ctx:       &context,

			before: func() {
				body := []Body{{
					X: 100,
					Y: stringVal,
				}}
				marshal, _ := yaml.Marshal(body)
				context.Request.Body = io.NopCloser(bytes.NewBuffer(marshal))
				context.Request.Header = http.Header{
					"Content-Type": []string{"application/x-test.yaml"},
				}
				injector.isInjectedBody = true
			},

			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Error(t, err)
				return false
			},

			bindErr: func(t assert.TestingT, err error) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.before != nil {
				tt.before()
			}
			elemT := reflect.TypeOf(req).Elem()
			ttx, b := elemT.FieldByName(tt.fieldName)
			assert.True(t, b)
			fmt.Printf("%v", ttx)

			fn, err := injector.inject(
				tt.kind,
				tt.key,
				ttx,
			)
			if tt.wantErr(t, err) {
				tt.bindErr(t, fn(tt.ctx, reflect.ValueOf(req).Elem()))
			}
		})
	}
}

func Test_httpInjector_BindFuncs(t *testing.T) {
	t.Run("Suc", func(t *testing.T) {
		injector := httpInjector{}
		injector.bindFuncs = append(injector.bindFuncs, func(ctx *gin.Context, v reflect.Value) error {
			field := fieldByIndexFromStructValue(v, []int{0}, false, reflect.TypeOf(1000))
			field.SetInt(1000)
			return nil
		})

		fn := injector.BindFuncs()
		injector.bindFuncs = nil

		type X struct {
			x int
			y int
		}
		req := X{y: 200}

		value, err := fn(&gin.Context{}, reflect.ValueOf(&req).Elem())
		assert.Nil(t, err)

		x := value.Interface().(X)
		assert.Equal(t, 1000, x.x)
		assert.Equal(t, 200, x.y)
	})
}

func Test_bitSize(t *testing.T) {
	type args struct {
		kind reflect.Kind
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "int64",
			args: args{
				kind: reflect.Int64,
			},
			want: 64,
		},
		{
			name: "int",
			args: args{
				kind: reflect.Int,
			},
			want: 32,
		},
		{
			name: "int16",
			args: args{
				kind: reflect.Int16,
			},
			want: 16,
		},
		{
			name: "int8",
			args: args{
				kind: reflect.Int8,
			},
			want: 8,
		},
		{
			name: "default",
			args: args{
				kind: reflect.Map,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, bitSize(tt.args.kind), "bitSize(%v)", tt.args.kind)
		})
	}
}
