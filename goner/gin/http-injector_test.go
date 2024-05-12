package gin

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/url"
	"testing"
)

var f func()

func (c *Context) Next() {
	f()
}

func Test_HttpInject(t *testing.T) {
	gone.
		Prepare(func(cemetery gone.Cemetery) error {
			_ = tracer.Priest(cemetery)
			_ = logrus.Priest(cemetery)
			_ = config.Priest(cemetery)
			cemetery.Bury(NewHttInjector())
			return nil
		}).
		AfterStart(func(in struct {
			log          logrus.Logger `gone:"gone-logger"`
			httpInjector httpInjector  `gone:"http"`
			cemetery     gone.Cemetery `gone:"gone-cemetery"`
			tracer       tracer.Tracer `gone:"gone-tracer"`
		}) {
			in.tracer.SetTraceId("", func() {

				type MockWriter struct {
					gin.ResponseWriter
				}

				Url, _ := url.Parse("https://goner.fun/zh/?page=1&pageSize=10&arr=1&arr=2&arr=3")

				context := Context{
					Context: &gin.Context{
						Request: &http.Request{
							URL: Url,
							Header: http.Header{
								"Host":   {"goner.fun"},
								"Cookie": {"key1=v1;key2=v2;"},
							},
						},
						Writer: &MockWriter{},
					},
				}

				var err error
				_, err = gone.InjectWrapFn(in.cemetery, func() {})
				assert.Nil(t, err)

				var i = 0
				f = func() {
					fn, err := gone.InjectWrapFn(in.cemetery, func(arg struct {
						ctx    Context            `gone:"http"`
						ctxP   *Context           `gone:"http"`
						req    http.Request       `gone:"http"`
						reqP   *http.Request      `gone:"http"`
						url    url.URL            `gone:"http"`
						urlP   *url.URL           `gone:"http"`
						writer gin.ResponseWriter `gone:"http"`
						header http.Header        `gone:"http"`
						log    logrus.Logger      `gone:"gone-logger"`
					}) {
						i++
						assert.Equal(t, context, arg.ctx)
						assert.Equal(t, &context, arg.ctxP)
						assert.Equal(t, *context.Request, arg.req)
						assert.Equal(t, context.Request, arg.reqP)
						assert.Equal(t, *context.Request.URL, arg.url)
						assert.Equal(t, context.Request.URL, arg.urlP)
						assert.Equal(t, context.Writer, arg.writer)
						assert.Equal(t, context.Request.Header, arg.header)
						assert.Equal(t, in.log, arg.log)
					})

					assert.Nil(t, err)
					assert.NotNil(t, fn)
					_ = gone.ExecuteInjectWrapFn(fn)
					assert.Equal(t, 1, i)
					i++
				}
				_, _ = in.httpInjector.SetContext(&context)
				assert.Equal(t, 2, i)

				//cannot inject
				i = 0
				f = func() {
					fn, err := gone.InjectWrapFn(in.cemetery, func(arg struct {
						m map[string]string `gone:"http"`
					}) {
						i++
					})

					assert.NotNil(t, err)
					assert.Nil(t, fn)
					i++
				}
				_, _ = in.httpInjector.SetContext(&context)
				assert.Equal(t, 1, i)

				//inject query,cookie,header
				i = 0
				f = func() {
					fn, err := gone.InjectWrapFn(in.cemetery, func(arg struct {
						page  int       `gone:"http,query"`
						page1 int       `gone:"http,query=page"`
						arr   []int     `gone:"http,query=arr"`
						arr2  []uint    `gone:"http,query=arr"`
						arr3  []int8    `gone:"http,query=arr"`
						arr4  []float64 `gone:"http,query=arr"`
						arr5  []float32 `gone:"http,query=arr"`
						h     string    `gone:"http,header=Host"`
						h2    string    `gone:"http,header=host"`
						c     string    `gone:"http,cookie=key1"`
					}) {
						i++
						assert.Equal(t, arg.page, 1)
						assert.Equal(t, arg.page1, 1)
						assert.Equal(t, len(arg.arr), 3)
						assert.Equal(t, arg.h, "goner.fun")
						assert.Equal(t, arg.h2, "goner.fun")
						assert.Equal(t, arg.c, "v1")
					})

					assert.NotNil(t, fn)
					assert.Nil(t, err)

					_ = gone.ExecuteInjectWrapFn(fn)
					in.log.Infof("%v", err)
					i++
				}
				_, _ = in.httpInjector.SetContext(&context)
				assert.Equal(t, 2, i)

				//inject body
				type Req struct {
					Test  string `form:"test"`
					Test1 int    `form:"test1"`
					Test2 bool   `form:"test2"`
				}
				req := Req{
					Test:  "test",
					Test1: 1,
					Test2: true,
				}

				marshal, _ := json.Marshal(req)

				context.Context.Request.Body = io.NopCloser(bytes.NewReader(marshal))
				context.Context.Request.Header.Set("Content-Type", "application/json")

				i = 0
				f = func() {
					fn, err := gone.InjectWrapFn(in.cemetery, func(arg struct {
						req Req `gone:"http,body"`
					}) {
						i++
						assert.Equal(t, arg.req.Test, "test")
						assert.Equal(t, arg.req.Test1, 1)
						assert.Equal(t, arg.req.Test2, true)
					})

					assert.Nil(t, err)
					assert.NotNil(t, fn)
					_ = gone.ExecuteInjectWrapFn(fn)
					in.log.Infof("%v", err)
					i++
				}
				_, _ = in.httpInjector.SetContext(&context)
				assert.Equal(t, 2, i)

				//use pointer
				context.Context.Request.Body = io.NopCloser(bytes.NewReader(marshal))
				i = 0
				f = func() {
					fn, err := gone.InjectWrapFn(in.cemetery, func(arg struct {
						req *Req `gone:"http,body"`
					}) {
						i++
						assert.Equal(t, arg.req.Test, "test")
						assert.Equal(t, arg.req.Test1, 1)
						assert.Equal(t, arg.req.Test2, true)
					})

					assert.Nil(t, err)
					assert.NotNil(t, fn)
					_ = gone.ExecuteInjectWrapFn(fn)
					in.log.Infof("%v", err)
					i++
				}
				_, _ = in.httpInjector.SetContext(&context)
				assert.Equal(t, 2, i)

				//use xml
				xmlBytes, _ := xml.Marshal(req)
				context.Context.Request.Body = io.NopCloser(bytes.NewReader(xmlBytes))
				context.Context.Request.Header.Set("Content-Type", "application/xml")
				i = 0
				f = func() {
					fn, err := gone.InjectWrapFn(in.cemetery, func(arg struct {
						req *Req `gone:"http,body"`
					}) {
						i++
						assert.Equal(t, arg.req.Test, "test")
						assert.Equal(t, arg.req.Test1, 1)
						assert.Equal(t, arg.req.Test2, true)
					})

					assert.Nil(t, err)
					assert.NotNil(t, fn)
					_ = gone.ExecuteInjectWrapFn(fn)
					in.log.Infof("%v", err)
					i++
				}
				_, _ = in.httpInjector.SetContext(&context)
				assert.Equal(t, 2, i)

				//use form

			})
		}).
		Run()
}

func Test_parseConfKeyValue(t *testing.T) {
	type args struct {
		conf string
	}
	tests := []struct {
		name      string
		args      args
		wantKey   string
		wantValue string
	}{
		{
			name: "x=100",
			args: args{
				conf: "x=100",
			},
			wantKey:   "x",
			wantValue: "100",
		},
		{
			name: "x=",
			args: args{
				conf: "x=",
			},
			wantKey:   "x",
			wantValue: "",
		},
		{
			name: "x",
			args: args{
				conf: "x",
			},
			wantKey:   "x",
			wantValue: "",
		},
		{
			name: "=111",
			args: args{
				conf: "=111",
			},
			wantKey:   "",
			wantValue: "111",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKey, gotValue := parseConfKeyValue(tt.args.conf)
			assert.Equalf(t, tt.wantKey, gotKey, "parseConfKeyValue(%v)", tt.args.conf)
			assert.Equalf(t, tt.wantValue, gotValue, "parseConfKeyValue(%v)", tt.args.conf)
		})
	}
}
