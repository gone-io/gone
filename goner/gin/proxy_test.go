package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func Test_proxy_proxyOne1(t *testing.T) {
	gone.
		Prepare(func(cemetery gone.Cemetery) error {
			_ = config.Priest(cemetery)
			_ = logrus.Priest(cemetery)
			_ = tracer.Priest(cemetery)
			cemetery.Bury(NewGinProxy())
			cemetery.Bury(NewHttInjector())
			cemetery.Bury(NewGinResponser())
			return nil
		}).
		Test(func(in struct {
			proxy HandleProxyToGin `gone:"*"`
		}) {
			controller := gomock.NewController(t)
			defer controller.Finish()
			writer := NewMockResponseWriter(controller)
			writer.EXPECT().Written().AnyTimes()
			writer.EXPECT().WriteHeader(gomock.Any()).AnyTimes()
			writer.EXPECT().Header().Return(http.Header{}).AnyTimes()
			writer.EXPECT().Write(gomock.Any()).AnyTimes()

			Url, _ := url.Parse("https://goner.fun/zh/?page=1&pageSize=10&arr=1&arr=2&arr=3")

			context := gin.Context{
				Writer: writer,
				Request: &http.Request{
					URL: Url,
				},
			}

			t.Run("ctx inject", func(t *testing.T) {
				executedCounter := 0
				proxyFn := in.proxy.Proxy(func(in struct {
					ctx    gin.Context  `gone:"http"`
					ctxPtr *gin.Context `gone:"http"`
				}) {
					assert.Equal(t, in.ctxPtr, &context)
					assert.NotNil(t, in.ctx.Writer, context.Writer)
					executedCounter++
					return
				})[0]
				proxyFn(&context)
				assert.Equal(t, executedCounter, 1)
			})

			t.Run("request inject", func(t *testing.T) {
				executedCounter := 0
				proxyFn := in.proxy.Proxy(func(in struct {
					req    http.Request  `gone:"http"`
					reqPtr *http.Request `gone:"http"`
				}) {
					assert.Equal(t, in.req.URL, context.Request.URL)
					assert.NotNil(t, in.reqPtr, context.Request)
					executedCounter++
					return
				})[0]
				proxyFn(&context)
				assert.Equal(t, executedCounter, 1)
			})

		})
}
