package gin

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

//func Test_HttpInject(t *testing.T) {
//
//	gone.
//		Prepare(func(cemetery gone.Cemetery) error {
//			_ = tracer.Priest(cemetery)
//			_ = logrus.Priest(cemetery)
//			_ = config.Priest(cemetery)
//			cemetery.Bury(NewHttInjector())
//			return nil
//		}).
//		AfterStart(func(in struct {
//			log          gone.Logger   `gone:"gone-logger"`
//			httpInjector httpInjector  `gone:"http"`
//			cemetery     gone.Cemetery `gone:"gone-cemetery"`
//			tracer       gone.Tracer   `gone:"gone-tracer"`
//		}) {
//			in.tracer.SetTraceId("", func() {
//
//				type MockWriter struct {
//					gin.ResponseWriter
//				}
//
//				Url, _ := url.Parse("https://goner.fun/zh/?page=1&pageSize=10&arr=1&arr=2&arr=3")
//
//				context := Context{
//					Context: &gin.Context{
//						Request: &http.Request{
//							URL: Url,
//							Header: http.Header{
//								"Host":   {"goner.fun"},
//								"Cookie": {"key1=v1;key2=v2;"},
//							},
//						},
//						Writer: &MockWriter{},
//					},
//				}
//
//				var err error
//				var f func()
//				_, err = gone.InjectWrapFn(in.cemetery, func() {})
//				assert.Nil(t, err)
//
//				var i = 0
//				f = func() {
//					fn, err := gone.InjectWrapFn(in.cemetery, func(arg struct {
//						ctx    Context            `gone:"http"`
//						ctxP   *Context           `gone:"http"`
//						req    http.Request       `gone:"http"`
//						reqP   *http.Request      `gone:"http"`
//						url    url.URL            `gone:"http"`
//						urlP   *url.URL           `gone:"http"`
//						writer gin.ResponseWriter `gone:"http"`
//						header http.Header        `gone:"http"`
//						log    logrus.Logger      `gone:"gone-logger"`
//					}) {
//						i++
//						assert.Equal(t, context, arg.ctx)
//						assert.Equal(t, &context, arg.ctxP)
//						assert.Equal(t, *context.Request, arg.req)
//						assert.Equal(t, context.Request, arg.reqP)
//						assert.Equal(t, *context.Request.URL, arg.url)
//						assert.Equal(t, context.Request.URL, arg.urlP)
//						assert.Equal(t, context.Writer, arg.writer)
//						assert.Equal(t, context.Request.Header, arg.header)
//						assert.Equal(t, in.log, arg.log)
//					})
//
//					assert.Nil(t, err)
//					assert.NotNil(t, fn)
//					_ = gone.ExecuteInjectWrapFn(fn)
//					assert.Equal(t, 1, i)
//					i++
//				}
//				_, _ = in.httpInjector.setContext(&context, f)
//				assert.Equal(t, 2, i)
//
//				//cannot inject
//				i = 0
//				f = func() {
//					fn, err := gone.InjectWrapFn(in.cemetery, func(arg struct {
//						m map[string]string `gone:"http"`
//					}) {
//						i++
//					})
//
//					assert.NotNil(t, err)
//					assert.Nil(t, fn)
//					i++
//				}
//				_, _ = in.httpInjector.setContext(&context, f)
//				assert.Equal(t, 1, i)
//
//				//inject query,cookie,header
//				i = 0
//				f = func() {
//					fn, err := gone.InjectWrapFn(in.cemetery, func(arg struct {
//						page  int       `gone:"http,query"`
//						page1 int       `gone:"http,query=page"`
//						arr   []int     `gone:"http,query=arr"`
//						arr2  []uint    `gone:"http,query=arr"`
//						arr3  []int8    `gone:"http,query=arr"`
//						arr4  []float64 `gone:"http,query=arr"`
//						arr5  []float32 `gone:"http,query=arr"`
//						h     string    `gone:"http,header=Host"`
//						h2    string    `gone:"http,header=host"`
//						c     string    `gone:"http,cookie=key1"`
//					}) {
//						i++
//						assert.Equal(t, arg.page, 1)
//						assert.Equal(t, arg.page1, 1)
//						assert.Equal(t, len(arg.arr), 3)
//						assert.Equal(t, arg.h, "goner.fun")
//						assert.Equal(t, arg.h2, "goner.fun")
//						assert.Equal(t, arg.c, "v1")
//					})
//
//					assert.NotNil(t, fn)
//					assert.Nil(t, err)
//
//					_ = gone.ExecuteInjectWrapFn(fn)
//					in.log.Infof("%v", err)
//					i++
//				}
//				_, _ = in.httpInjector.setContext(&context, f)
//				assert.Equal(t, 2, i)
//
//				//inject body
//				type Req struct {
//					Test  string `form:"test"`
//					Test1 int    `form:"test1"`
//					Test2 bool   `form:"test2"`
//				}
//				req := Req{
//					Test:  "test",
//					Test1: 1,
//					Test2: true,
//				}
//
//				marshal, _ := json.Marshal(req)
//
//				context.Context.Request.Body = io.NopCloser(bytes.NewReader(marshal))
//				context.Context.Request.Header.Set("Content-Type", "application/json")
//
//				i = 0
//				f = func() {
//					fn, err := gone.InjectWrapFn(in.cemetery, func(arg struct {
//						req Req `gone:"http,body"`
//					}) {
//						i++
//						assert.Equal(t, arg.req.Test, "test")
//						assert.Equal(t, arg.req.Test1, 1)
//						assert.Equal(t, arg.req.Test2, true)
//					})
//
//					assert.Nil(t, err)
//					assert.NotNil(t, fn)
//					_ = gone.ExecuteInjectWrapFn(fn)
//					in.log.Infof("%v", err)
//					i++
//				}
//				_, _ = in.httpInjector.setContext(&context, f)
//				assert.Equal(t, 2, i)
//
//				//use pointer
//				context.Context.Request.Body = io.NopCloser(bytes.NewReader(marshal))
//				i = 0
//				f = func() {
//					fn, err := gone.InjectWrapFn(in.cemetery, func(arg struct {
//						req *Req `gone:"http,body"`
//					}) {
//						i++
//						assert.Equal(t, arg.req.Test, "test")
//						assert.Equal(t, arg.req.Test1, 1)
//						assert.Equal(t, arg.req.Test2, true)
//					})
//
//					assert.Nil(t, err)
//					assert.NotNil(t, fn)
//					_ = gone.ExecuteInjectWrapFn(fn)
//					in.log.Infof("%v", err)
//					i++
//				}
//				_, _ = in.httpInjector.setContext(&context, f)
//				assert.Equal(t, 2, i)
//
//				//use xml
//				xmlBytes, _ := xml.Marshal(req)
//				context.Context.Request.Body = io.NopCloser(bytes.NewReader(xmlBytes))
//				context.Context.Request.Header.Set("Content-Type", "application/xml")
//				i = 0
//				f = func() {
//					fn, err := gone.InjectWrapFn(in.cemetery, func(arg struct {
//						req *Req `gone:"http,body"`
//					}) {
//						i++
//						assert.Equal(t, arg.req.Test, "test")
//						assert.Equal(t, arg.req.Test1, 1)
//						assert.Equal(t, arg.req.Test2, true)
//					})
//
//					assert.Nil(t, err)
//					assert.NotNil(t, fn)
//					_ = gone.ExecuteInjectWrapFn(fn)
//					in.log.Infof("%v", err)
//					i++
//				}
//				_, _ = in.httpInjector.setContext(&context, f)
//				assert.Equal(t, 2, i)
//
//				//use form
//
//			})
//		}).
//		Run()
//}

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

func (s *httpInjector) Errorf(format string, args ...any) {}

func Test_httpInjector_injectBody(t *testing.T) {
	t.Run("unsupportedAttributeType", func(t *testing.T) {
		var test = "test"
		s := &httpInjector{}
		err := s.injectBody(reflect.ValueOf(test), "test")
		assert.Error(t, err)
	})

	//t.Run("NewParameterError", func(t *testing.T) {
	//	type Req struct {
	//		X string `json:"x,omitempty"`
	//		Y string `json:"y,omitempty" binding:"required"`
	//	}
	//	var req = Req{X: "test"}
	//	marshal, _ := json.Marshal(req)
	//
	//	context := gin.Context{
	//		Request: &http.Request{
	//			Body: io.NopCloser(bytes.NewReader(marshal)),
	//			Header: http.Header{
	//				"Content-Type": []string{"application/json"},
	//			},
	//		},
	//	}
	//
	//	var InjectAgs *struct {
	//		Req Req
	//	}
	//
	//	s := &httpInjector{}
	//	err := s.injectBody(reflect.ValueOf(InjectAgs.Req), "req")
	//	assert.Error(t, err)
	//	funcs := s.CollectBindFuncs()
	//	err = funcs[0](&context)
	//	assert.Nil(t, err)
	//
	//	err = s.injectBody(reflect.ValueOf(&req), "req")
	//	assert.Error(t, err)
	//})
}

//
//func Test_httpInjector_injectByKind(t *testing.T) {
//	t.Run("kind=param", func(t *testing.T) {
//		context := Context{
//			Context: &gin.Context{
//				Params: gin.Params{
//					gin.Param{
//						Key:   "test",
//						Value: "100",
//					},
//				},
//			},
//		}
//
//		type Req struct {
//			X string
//		}
//
//		var req = &Req{X: "test"}
//
//		s := &httpInjector{}
//		err := s.injectByKind(&context, keyParam, "test", reflect.ValueOf(req).Elem().FieldByName("X"), "test")
//		assert.Nil(t, err)
//		assert.Equal(t, "100", req.X)
//	})
//
//	t.Run("kind=cookie", func(t *testing.T) {
//		context := Context{
//			Context: &gin.Context{
//				Request: &http.Request{
//					Header: http.Header{
//						"Host":   {"goner.fun"},
//						"Cookie": {"key1=v1;key2=v2;"},
//					},
//				},
//			},
//		}
//
//		type Req struct {
//			X string
//		}
//
//		var req = &Req{X: "test"}
//
//		s := &httpInjector{}
//		err := s.injectByKind(&context, keyCookie, "test", reflect.ValueOf(req).Elem().FieldByName("X"), "test")
//		assert.Error(t, err)
//	})
//
//	t.Run("unsupportedKindConfigure", func(t *testing.T) {
//		context := Context{}
//
//		type Req struct {
//			X string
//		}
//
//		var req = &Req{X: "test"}
//
//		s := &httpInjector{}
//		err := s.injectByKind(&context, "other", "test", reflect.ValueOf(req).Elem().FieldByName("X"), "test")
//		assert.Error(t, err)
//	})
//}
//
//func Test_httpInjector_injectQuery(t *testing.T) {
//	Url, _ := url.Parse("https://goner.fun/zh/?page=1&pageSize=10&arr=1&arr=2&arr=3")
//
//	type Req1 struct {
//		Page     int   `form:"page"`
//		PageSize int   `form:"pageSize"`
//		Arr      []int `form:"arr"`
//	}
//	type Req2 struct {
//		Page     int `form:"page"`
//		PageSize int `form:"pageSize"`
//		X        int `form:"x" binding:"required"`
//	}
//
//	t.Run("typ=struct", func(t *testing.T) {
//		context := Context{
//			Context: &gin.Context{
//				Request: &http.Request{
//					URL: Url,
//				},
//			},
//		}
//		s := &httpInjector{}
//		t.Run("struct suc", func(t *testing.T) {
//			type Q struct {
//				X Req1
//			}
//
//			var req = &Q{}
//
//			err := s.injectQuery(&context, reflect.ValueOf(req).Elem().FieldByName("X"), "test", "")
//			assert.Nil(t, err)
//			assert.Equal(t, 1, req.X.Page)
//			assert.Equal(t, 10, req.X.PageSize)
//		})
//
//		t.Run("struct err", func(t *testing.T) {
//			type Q struct {
//				X Req2
//			}
//			var req = &Q{}
//			err := s.injectQuery(&context, reflect.ValueOf(req).Elem().FieldByName("X"), "test", "")
//			assert.Error(t, err)
//		})
//
//		t.Run("struct pointer suc", func(t *testing.T) {
//			type Q struct {
//				X *Req1
//			}
//
//			var req = &Q{}
//
//			err := s.injectQuery(&context, reflect.ValueOf(req).Elem().FieldByName("X"), "test", "")
//			assert.Nil(t, err)
//			assert.Equal(t, 1, req.X.Page)
//			assert.Equal(t, 10, req.X.PageSize)
//		})
//
//		t.Run("struct pointer err", func(t *testing.T) {
//			type Q struct {
//				X *Req2
//			}
//			var req = &Q{}
//			err := s.injectQuery(&context, reflect.ValueOf(req).Elem().FieldByName("X"), "test", "")
//			assert.Error(t, err)
//		})
//	})
//}
//
//func Test_httpInjector_injectQueryArray(t *testing.T) {
//	Url, _ := url.Parse("https://goner.fun/zh/?page=1&pageSize=10&arr=1&arr=2&arr=3")
//
//	context := Context{
//		Context: &gin.Context{
//			Request: &http.Request{
//				URL: Url,
//			},
//		},
//	}
//	s := &httpInjector{}
//	t.Run("[]string", func(t *testing.T) {
//		type Req struct {
//			Arr []string
//		}
//		var req = &Req{}
//
//		err := s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr"), "")
//		assert.Nil(t, err)
//		assert.Equal(t, []string{"1", "2", "3"}, req.Arr)
//	})
//	t.Run("[]int", func(t *testing.T) {
//		type Req struct {
//			Arr []int
//		}
//		var req = &Req{}
//		err := s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr"), "")
//		assert.Nil(t, err)
//		assert.Equal(t, []int{1, 2, 3}, req.Arr)
//	})
//	t.Run("[]int64", func(t *testing.T) {
//		type Req struct {
//			Arr []int64
//		}
//		var req = &Req{}
//		err := s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr"), "")
//		assert.Nil(t, err)
//		assert.Equal(t, []int64{1, 2, 3}, req.Arr)
//	})
//	t.Run("[]int32", func(t *testing.T) {
//		type Req struct {
//			Arr []int32
//		}
//		var req = &Req{}
//		err := s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr"), "")
//		assert.Nil(t, err)
//		assert.Equal(t, []int32{1, 2, 3}, req.Arr)
//	})
//	t.Run("[]int16", func(t *testing.T) {
//		type Req struct {
//			Arr []int16
//		}
//		var req = &Req{}
//		err := s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr"), "")
//		assert.Nil(t, err)
//		assert.Equal(t, []int16{1, 2, 3}, req.Arr)
//	})
//	t.Run("[]int8", func(t *testing.T) {
//		type Req struct {
//			Arr []int8
//		}
//		var req = &Req{}
//		err := s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr"), "")
//		assert.Nil(t, err)
//		assert.Equal(t, []int8{1, 2, 3}, req.Arr)
//	})
//
//	t.Run("[]uint", func(t *testing.T) {
//		type Req struct {
//			Arr []uint
//		}
//		var req = &Req{}
//		err := s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr"), "")
//		assert.Nil(t, err)
//		assert.Equal(t, []uint{1, 2, 3}, req.Arr)
//	})
//
//	t.Run("[]uint64", func(t *testing.T) {
//		type Req struct {
//			Arr []uint64
//		}
//		var req = &Req{}
//		err := s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr"), "")
//		assert.Nil(t, err)
//		assert.Equal(t, []uint64{1, 2, 3}, req.Arr)
//	})
//	t.Run("[]uint32", func(t *testing.T) {
//		type Req struct {
//			Arr []uint32
//		}
//		var req = &Req{}
//		err := s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr"), "")
//		assert.Nil(t, err)
//		assert.Equal(t, []uint32{1, 2, 3}, req.Arr)
//	})
//	t.Run("[]uint16", func(t *testing.T) {
//		type Req struct {
//			Arr []uint16
//		}
//		var req = &Req{}
//		err := s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr"), "")
//		assert.Nil(t, err)
//		assert.Equal(t, []uint16{1, 2, 3}, req.Arr)
//	})
//	t.Run("[]uint8", func(t *testing.T) {
//		type Req struct {
//			Arr []uint8
//		}
//		var req = &Req{}
//		err := s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr"), "")
//		assert.Nil(t, err)
//		assert.Equal(t, []uint8{1, 2, 3}, req.Arr)
//	})
//
//	t.Run("[]bool", func(t *testing.T) {
//		type Req struct {
//			Arr []bool
//		}
//		var req = &Req{}
//		err := s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr"), "")
//		assert.Nil(t, err)
//		assert.Equal(t, []bool{true, true, true}, req.Arr)
//	})
//
//}
//
//func Test_xx(t *testing.T) {
//	t.Run("err", func(t *testing.T) {
//		Url, _ := url.Parse("https://goner.fun/zh/?page=1&pageSize=10&arr=1&arr=2&arr=i3")
//
//		context := Context{
//			Context: &gin.Context{
//				Request: &http.Request{
//					URL: Url,
//				},
//			},
//		}
//		s := &httpInjector{}
//
//		type Req struct {
//			Arr1 []int
//			Arr2 []uint
//			Arr3 []float64
//			Arr4 []func()
//		}
//		var req = &Req{}
//		err := s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr1"), "")
//		assert.Error(t, err)
//
//		err = s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr2"), "")
//		assert.Error(t, err)
//
//		err = s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr3"), "")
//		assert.Error(t, err)
//
//		err = s.injectQueryArray(&context, "arr", reflect.ValueOf(req).Elem().FieldByName("Arr4"), "")
//		assert.Error(t, err)
//	})
//}
//
//func Test_httpInjector_parseStringValueAndInject(t *testing.T) {
//	s := &httpInjector{}
//	t.Run("bool", func(t *testing.T) {
//		type Req struct {
//			X bool
//		}
//		var req = &Req{}
//		err := s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X"), "test", "true")
//		assert.Nil(t, err)
//		assert.True(t, req.X)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X"), "test", "")
//		assert.Nil(t, err)
//		assert.False(t, req.X)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X"), "test", "false")
//		assert.Nil(t, err)
//		assert.False(t, req.X)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X"), "test", "0")
//		assert.Nil(t, err)
//		assert.False(t, req.X)
//	})
//
//	t.Run("int", func(t *testing.T) {
//		type Req struct {
//			X1 int
//			X2 int64
//			X3 int32
//			X4 int16
//			X5 int8
//		}
//		var req = &Req{}
//		var err error
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X1"), "test", "100")
//		assert.Nil(t, err)
//		assert.Equal(t, int(100), req.X1)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X2"), "test", "100")
//		assert.Nil(t, err)
//		assert.Equal(t, int64(100), req.X2)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X3"), "test", "100")
//		assert.Nil(t, err)
//		assert.Equal(t, int32(100), req.X3)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X4"), "test", "100")
//		assert.Nil(t, err)
//		assert.Equal(t, int16(100), req.X4)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X5"), "test", "100")
//		assert.Nil(t, err)
//		assert.Equal(t, int8(100), req.X5)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X1"), "test", "x100")
//		assert.Error(t, err)
//	})
//
//	t.Run("uint", func(t *testing.T) {
//		type Req struct {
//			X1 uint
//			X2 uint64
//			X3 uint32
//			X4 uint16
//			X5 uint8
//		}
//		var req = &Req{}
//		var err error
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X1"), "test", "100")
//		assert.Nil(t, err)
//		assert.Equal(t, uint(100), req.X1)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X2"), "test", "100")
//		assert.Nil(t, err)
//		assert.Equal(t, uint64(100), req.X2)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X3"), "test", "100")
//		assert.Nil(t, err)
//		assert.Equal(t, uint32(100), req.X3)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X4"), "test", "100")
//		assert.Nil(t, err)
//		assert.Equal(t, uint16(100), req.X4)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X5"), "test", "100")
//		assert.Nil(t, err)
//		assert.Equal(t, uint8(100), req.X5)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X1"), "test", "x100")
//		assert.Error(t, err)
//	})
//
//	t.Run("float", func(t *testing.T) {
//		type Req struct {
//			X1 float32
//			X2 float64
//		}
//		var req = &Req{}
//		var err error
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X1"), "test", "100.1")
//		assert.Nil(t, err)
//		assert.Equal(t, float32(100.1), req.X1)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X2"), "test", "100.1")
//		assert.Nil(t, err)
//		assert.Equal(t, float64(100.1), req.X2)
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X1"), "test", "x100.1")
//		assert.Error(t, err)
//	})
//
//	t.Run("unsigned", func(t *testing.T) {
//		type Req struct {
//			X1 func()
//		}
//
//		var req = &Req{}
//		var err error
//
//		err = s.parseStringValueAndInject(reflect.ValueOf(req).Elem().FieldByName("X1"), "test", "100.1")
//		assert.Error(t, err)
//	})
//}
//
//func Test_httpInjector_SetContext(t *testing.T) {
//	gone.Prepare(tracer.Priest, logrus.Priest, config.Priest, func(cemetery gone.Cemetery) error {
//		cemetery.Bury(NewHttInjector())
//		return nil
//	}).AfterStart(func(in struct {
//		inject *httpInjector `gone:"*"`
//		tracer gone.Tracer   `gone:"*"`
//	}) {
//		in.tracer.SetTraceId("", func() {
//			defer func() {
//				a := recover()
//				assert.NotNil(t, a)
//			}()
//
//			_, err := in.inject.SetContext(&Context{})
//			assert.Nil(t, err)
//		})
//	}).Run()
//}

func Test_httpInjector_Suck(t *testing.T) {

	type Req struct {
		X string
	}
	var req = &Req{}

	field, b := reflect.TypeOf(req).Elem().FieldByName("X")
	assert.True(t, b)

	gone.Prepare(tracer.Priest, logrus.Priest, config.Priest, func(cemetery gone.Cemetery) error {
		cemetery.Bury(NewHttInjector())
		return nil
	}).AfterStart(func(in struct {
		inject *httpInjector `gone:"*"`
		tracer gone.Tracer   `gone:"*"`
	}) {
		err := in.inject.Suck("x", reflect.ValueOf(req).Elem().FieldByName("X"), field)
		assert.Error(t, err)

		in.tracer.SetTraceId("", func() {
			err := in.inject.Suck("x", reflect.ValueOf(req).Elem().FieldByName("X"), field)
			assert.Error(t, err)
		})
	}).Run()
}
