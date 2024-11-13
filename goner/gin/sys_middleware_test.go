package gin

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/internal/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"net/url"
	"testing"
)

//
//func (p *sysProcessor) Errorf(format string, args ...any) {}
//func (p *sysProcessor) Warnf(format string, args ...any)  {}
//func (p *sysProcessor) Infof(format string, args ...any)  {}
//
//func Test_sysProcessor_AfterRevive(t *testing.T) {
//	controller := gomock.NewController(t)
//	defer controller.Finish()
//
//	t.Run("ShowRequestTime", func(t *testing.T) {
//		iRouter := NewMockIRouter(controller)
//		iRouter.EXPECT().Use(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
//
//		processor := sysProcessor{
//			router:          iRouter,
//			ShowRequestTime: true,
//			ShowAccessLog:   true,
//		}
//		_ = processor.AfterRevive()
//	})
//}
//
//type testTracer struct {
//	gone.Tracer
//}
//
//func (t *testTracer) SetTraceId(traceId string, fn func()) {
//	if traceId == "" {
//		traceId = "1234567890"
//	}
//	fn()
//}
//
//func (t *testTracer) GetTraceId() string {
//	return "1234567890"
//}
//
//func Test_sysProcessor_trace(t *testing.T) {
//	controller := gomock.NewController(t)
//	defer controller.Finish()
//
//	t.Run("HealthCheckUrl", func(t *testing.T) {
//		writer := NewMockResponseWriter(controller)
//		writer.EXPECT().WriteHeader(200)
//		writer.EXPECT().WriteHeaderNow()
//
//		processor := sysProcessor{
//			HealthCheckUrl: "/health",
//		}
//
//		Url, _ := url.Parse("https://goner.fun/health")
//
//		context := Context{
//			Context: &gin.Context{
//				Request: &http.Request{
//					URL: Url,
//				},
//				Writer: writer,
//			},
//		}
//
//		_, err := processor.trace(&context)
//		assert.Nil(t, err)
//	})
//
//	t.Run("trace", func(t *testing.T) {
//		processor := sysProcessor{
//			tracer: &testTracer{},
//		}
//
//		Url, _ := url.Parse("https://goner.fun/health")
//
//		context := Context{
//			Context: &gin.Context{
//				Request: &http.Request{
//					URL: Url,
//				},
//			},
//		}
//
//		_, err := processor.trace(&context)
//		assert.Nil(t, err)
//	})
//}
//
//func Test_sysProcessor(t *testing.T) {
//	processor := sysProcessor{
//		tracer: &testTracer{},
//	}
//
//	Url, _ := url.Parse("https://goner.fun/health")
//
//	context := Context{
//		Context: &gin.Context{
//			Request: &http.Request{
//				URL: Url,
//			},
//		},
//	}
//
//	t.Run("recovery", func(t *testing.T) {
//		_, err := processor.recovery(&context)
//		assert.Nil(t, err)
//	})
//
//	t.Run("statRequestTime", func(t *testing.T) {
//		_, err := processor.statRequestTime(&context)
//		assert.Nil(t, err)
//	})
//
//	t.Run("accessLog", func(t *testing.T) {
//		controller := gomock.NewController(t)
//		defer controller.Finish()
//
//		writer := NewMockResponseWriter(controller)
//		writer.EXPECT().Header().Return(http.Header{})
//		writer.EXPECT().Status().Return(200)
//
//		context.Context.Writer = writer
//
//		context.Context.Request.Body = io.NopCloser(strings.NewReader("test"))
//
//		_, err := processor.accessLog(&context)
//		assert.Nil(t, err)
//	})
//
//	t.Run("accessLog-logDataMaxLength", func(t *testing.T) {
//		controller := gomock.NewController(t)
//		defer controller.Finish()
//
//		writer := NewMockResponseWriter(controller)
//		writer.EXPECT().Header().Return(http.Header{})
//		writer.EXPECT().Status().Return(200)
//		processor.logDataMaxLength = 2
//
//		context.Context.Writer = writer
//
//		context.Context.Request.Body = io.NopCloser(strings.NewReader("test"))
//
//		_, err := processor.accessLog(&context)
//		assert.Nil(t, err)
//	})
//
//	t.Run("accessLog-json", func(t *testing.T) {
//		controller := gomock.NewController(t)
//		defer controller.Finish()
//
//		writer := NewMockResponseWriter(controller)
//		writer.EXPECT().Header().Return(http.Header{
//			"Content-Type": {"application/json"},
//		})
//		writer.EXPECT().Status().Return(200)
//		processor.logDataMaxLength = 2
//
//		context.Context.Writer = writer
//
//		context.Context.Request.Body = io.NopCloser(strings.NewReader("test"))
//
//		_, err := processor.accessLog(&context)
//		assert.Nil(t, err)
//	})
//}
//
//func Test_sysProcessor_recover(t *testing.T) {
//	controller := gomock.NewController(t)
//	defer controller.Finish()
//	mockResponser := NewMockResponser(controller)
//
//	processor := sysProcessor{
//		tracer:     &testTracer{},
//		resHandler: mockResponser,
//	}
//
//	Url, _ := url.Parse("https://goner.fun/health")
//
//	context := Context{
//		Context: &gin.Context{
//			Request: &http.Request{
//				URL: Url,
//			},
//		},
//	}
//
//	mockResponser.EXPECT().Failed(gomock.Any(), gomock.Any()).Do(func(ctx any, err gone.InnerError) {
//		assert.Equal(t, 500, err.Code())
//	})
//
//	func() {
//		defer processor.recover(&context)
//		panic("err")
//	}()
//
//	assert.Nil(t, nil)
//}

func TestCustomResponseWrite(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	writer := NewMockResponseWriter(controller)
	writer.EXPECT().Write(gomock.Any()).Return(100, nil)
	writer.EXPECT().WriteString(gomock.Any()).Return(100, nil)

	blw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: writer}

	write, err := blw.Write([]byte("test"))
	assert.Equal(t, 100, write)
	assert.Nil(t, err)

	write, err = blw.WriteString("test2")
	assert.Equal(t, 100, write)
	assert.Nil(t, err)

	assert.Equal(t, "testtest2", blw.body.String())
}

func TestSysMiddleware_AfterRevive(t *testing.T) {
	type fields struct {
		enableLimit bool
		limit       rate.Limit
		burst       int
		limiter     *rate.Limiter
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "test",
			fields: fields{
				enableLimit: true,
				limit:       2,
				burst:       10,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Nil(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SysMiddleware{
				enableLimit: tt.fields.enableLimit,
				limit:       tt.fields.limit,
				burst:       tt.fields.burst,
				limiter:     tt.fields.limiter,
			}
			tt.wantErr(t, m.AfterRevive(), fmt.Sprintf("AfterRevive()"))
		})
	}
}

func TestSysMiddleware_allow(t *testing.T) {
	type fields struct {
		enableLimit bool
		limit       rate.Limit
		burst       int
		limiter     *rate.Limiter
	}
	tests := []struct {
		name   string
		fields fields
		before func(*SysMiddleware)
		want   bool
	}{
		{
			name: "test",
			fields: fields{
				enableLimit: true,
				limit:       2,
				burst:       10,
			},
			before: func(middleware *SysMiddleware) {
				err := middleware.AfterRevive()
				assert.Nil(t, err)
			},
			want: true,
		},
		{
			name: "test",
			fields: fields{
				enableLimit: true,
				limit:       0,
				burst:       0,
			},
			before: func(middleware *SysMiddleware) {
				err := middleware.AfterRevive()
				assert.Nil(t, err)
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SysMiddleware{
				enableLimit: tt.fields.enableLimit,
				limit:       tt.fields.limit,
				burst:       tt.fields.burst,
				limiter:     tt.fields.limiter,
			}
			if tt.before != nil {
				tt.before(m)
			}
			assert.Equalf(t, tt.want, m.allow(), "allow()")
		})
	}
}

func TestSysMiddleware_Process(t *testing.T) {
	t.Run("healthCheckUrl", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		writer := NewMockResponseWriter(controller)
		writer.EXPECT().WriteHeader(http.StatusOK)
		writer.EXPECT().WriteHeaderNow()

		Url, _ := url.Parse("/health")
		context := gin.Context{
			Request: &http.Request{
				URL: Url,
			},
			Writer: writer,
		}

		middleware := SysMiddleware{
			healthCheckUrl: "/health",
		}

		middleware.Process(&context)
	})

	t.Run("limited", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		xHeader := http.Header{}
		writer := NewMockResponseWriter(controller)
		writer.EXPECT().WriteHeader(http.StatusTooManyRequests)
		writer.EXPECT().Header().Return(xHeader)
		writer.EXPECT().Write(gomock.Any())

		context := gin.Context{
			Writer: writer,
		}
		middleware := SysMiddleware{}

		gone.Prepare(func(cemetery gone.Cemetery) error {
			cemetery.Bury(&middleware)
			cemetery.Bury(NewGinResponser())
			return logrus.Priest(cemetery)
		}).Test(func(middleware *SysMiddleware) {

			middleware.enableLimit = true
			middleware.limiter = rate.NewLimiter(0, 0)

			middleware.Process(&context)
		})
	})

	t.Run("use-tracer", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		h := http.Header{}
		traceId := uuid.New().String()
		h.Set(gone.TraceIdHeaderKey, traceId)
		context := gin.Context{
			Request: &http.Request{
				Header: h,
			},
		}
		middleware := SysMiddleware{}

		gone.Prepare(func(cemetery gone.Cemetery) error {
			cemetery.Bury(&middleware)
			cemetery.Bury(NewGinResponser())
			return logrus.Priest(cemetery)
		}).Test(func(middleware *SysMiddleware, tracer gone.Tracer) {
			middleware.showRequestLog = false
			middleware.showResponseLog = false
			middleware.showRequestTime = false

			var xTraceId string
			testInProcess = func(context *gin.Context) {
				xTraceId = tracer.GetTraceId()

			}
			middleware.Process(&context)
			assert.Equal(t, traceId, xTraceId)

			middleware.useTracer = false
			middleware.Process(&context)
			assert.Equal(t, "", xTraceId)
		})
	})
}

func TestSysMiddleware_requestLog(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	gone.Prepare(func(cemetery gone.Cemetery) error {
		cemetery.Bury(&SysMiddleware{})
		cemetery.Bury(NewGinResponser())
		return logrus.Priest(cemetery)
	}).Test(func(middleware *SysMiddleware, tracer gone.Tracer) {
		t.Run("X", func(t *testing.T) {
			logger := NewMockLogger(controller)
			logger.EXPECT().Infof(gomock.Any(), gomock.All()).Do(func(format string, args ...any) {
				assert.Equal(t, "[%s] %s", format)
				assert.Equal(t, "request", args[0])
				s := args[1].(string)

				assert.Contains(t, s, "referer")
				assert.Contains(t, s, "body")
				assert.Contains(t, s, "request-id")
				assert.Contains(t, s, "ip")
				assert.Contains(t, s, "method=POST")
				assert.Contains(t, s, "path=/health")
				assert.Contains(t, s, "user-agent")
			})

			middleware.logger = logger
			middleware.logDataMaxLength = 5

			Url, _ := url.Parse("http://localhost/health")
			xHeader := http.Header{}
			xHeader.Set("content-type", "application/json")
			context := gin.Context{
				Request: &http.Request{
					URL:        Url,
					Body:       io.NopCloser(bytes.NewBufferString("{\"data\":\"ok\"}")),
					Header:     xHeader,
					RemoteAddr: "127.0.0.1",
					Method:     "POST",
				},
			}
			middleware.requestLog(&context)
		})
	})
}

func TestSysMiddleware_responseLog(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	gone.Prepare(func(cemetery gone.Cemetery) error {
		cemetery.Bury(&SysMiddleware{})
		cemetery.Bury(NewGinResponser())
		return logrus.Priest(cemetery)
	}).Test(func(middleware *SysMiddleware, tracer gone.Tracer) {
		t.Run("X", func(t *testing.T) {
			logger := NewMockLogger(controller)
			logger.EXPECT().Infof(gomock.Any(), gomock.All()).Do(func(format string, args ...any) {
				assert.Equal(t, "[%s] %s", format)
				assert.Equal(t, "response", args[0])
				s := args[1].(string)

				assert.Contains(t, s, "body")
				assert.Contains(t, s, "method=POST")
				assert.Contains(t, s, "path=/health")
				assert.Contains(t, s, "content-type")
				assert.Contains(t, s, "status=200")
			})

			writer := NewMockResponseWriter(controller)
			writer.EXPECT().WriteHeader(http.StatusOK)
			xHeader := http.Header{}
			writer.EXPECT().Header().Return(xHeader).AnyTimes()
			writer.EXPECT().Write(gomock.Any())
			writer.EXPECT().Status().Return(http.StatusOK)

			middleware.logger = logger
			middleware.logDataMaxLength = 5

			Url, _ := url.Parse("http://localhost/health")
			context := gin.Context{
				Request: &http.Request{
					URL:    Url,
					Method: "POST",
				},
				Writer: writer,
			}
			middleware.responseLog(&context, func() {
				context.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok"})
			})
		})
	})
}

func TestSysMiddleware_log(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	t.Run("json", func(t *testing.T) {
		logger := NewMockLogger(controller)
		logger.EXPECT().Infof(gomock.Any(), gomock.All()).Do(func(format string, args ...any) {
			assert.Equal(t, "%s", format)
			b := args[0].([]byte)

			m := make(map[string]any)
			err := json.Unmarshal(b, &m)
			assert.Nil(t, err)
			assert.Equal(t, "ok", m["data"])
			assert.Equal(t, "test", m["type"])
		})

		middleware := SysMiddleware{logger: logger, logFormat: "json"}

		middleware.log("test", map[string]any{
			"data": "ok",
		})
	})
	t.Run("console", func(t *testing.T) {
		logger := NewMockLogger(controller)
		logger.EXPECT().Infof(gomock.Any(), gomock.All()).Do(func(format string, args ...any) {
			assert.Equal(t, "[%s] %s", format)
			assert.Equal(t, "test", args[0])
			s := args[1].(string)
			assert.Contains(t, s, "data")
		})

		middleware := SysMiddleware{logger: logger, logFormat: "console"}

		middleware.log("test", map[string]any{
			"data": "ok",
		})
	})

}
