package gin_test

import (
	"github.com/golang/mock/gomock"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/gin"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_proxy_Proxy(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	responser := gin.NewMockResponser(controller)
	injector := gin.NewMockHttInjector(controller)
	responser.EXPECT().Success(gomock.Any(), gomock.Any()).AnyTimes()
	responser.EXPECT().ProcessResults(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.All()).AnyTimes()

	gone.
		Prepare(config.Priest, tracer.Priest, logrus.Priest, func(cemetery gone.Cemetery) error {

			cemetery.Bury(gin.NewGinProxy())
			cemetery.Bury(responser)
			cemetery.Bury(injector)
			return nil
		}).
		Test(func(proxy gin.HandleProxyToGin, logger gone.Logger) {
			i := 0
			t.Run("Special funcs", func(t *testing.T) {
				funcs := proxy.Proxy(
					func(*gone.Context) (any, error) {
						i++
						return nil, nil
					},
					func(*gone.Context) error {
						i++
						return nil
					},
					func(*gone.Context) {
						i++
					},
					func(*gin.OriginContent) (any, error) {
						i++
						return nil, nil
					},
					func(*gin.OriginContent) error {
						i++
						return nil
					},
					func(*gin.OriginContent) {
						i++
					},
					func() {
						i++
					},
					func() (any, error) {
						i++
						return nil, nil
					},
					func() error {
						i++
						return nil
					},
				)
				for _, fn := range funcs {
					fn(&gin.OriginContent{})
				}

				assert.Equal(t, 9, i)
			})

			t.Run("Inject funcs success", func(t *testing.T) {
				i := 0

				type One struct {
					X1  string
					log gone.Logger `gone:"*"`
				}

				type Two struct {
					X2  string
					log gone.Logger `gone:"*"`
				}

				injector.EXPECT().StartBindFuncs().MinTimes(3).MaxTimes(3)

				injector.EXPECT().BindFuncs().Return(func(ctx *gin.OriginContent, obj any, T reflect.Type) (reflect.Value, error) {
					one, ok := obj.(One)
					assert.True(t, ok)
					assert.Equal(t, logger, one.log)

					one.X1 = "one"
					return reflect.ValueOf(one), nil
				})

				fn2 := func(ctx *gin.OriginContent, obj any, arg reflect.Type) (reflect.Value, error) {
					two, ok := obj.(Two)
					assert.True(t, ok)
					assert.Equal(t, logger, two.log)

					two.X2 = "two"
					return reflect.ValueOf(two), nil
				}

				injector.EXPECT().BindFuncs().Return(fn2)

				funcs := proxy.Proxy(func(
					one One,
					two Two,
					logger gone.Logger,

					ctxPtr *gone.Context,
					ctx gone.Context,
					ginCtxPtr *gin.OriginContent,
					ginCtx gin.OriginContent,
				) (any, any, any, any, any, error, int) {
					assert.NotNil(t, logger)
					assert.Equal(t, logger, one.log)
					assert.Equal(t, logger, two.log)
					assert.Equal(t, "one", one.X1)
					assert.Equal(t, "two", two.X2)

					assert.NotNil(t, ctxPtr)
					assert.Equal(t, *ctxPtr, ctx)
					assert.Equal(t, ctx.Context, ginCtxPtr)
					assert.Equal(t, *ctx.Context, ginCtx)
					i++
					var x *int = nil
					type X struct{}
					var s []int
					var s2 = make([]int, 0)
					return 10, s, s2, X{}, x, nil, 0
				})
				funcs[0](&gin.OriginContent{})
				assert.Equal(t, 1, i)
			})

			t.Run("Inject Error", func(t *testing.T) {
				defer func() {
					err := recover()
					assert.Error(t, err.(error))
				}()

				injector.EXPECT().StartBindFuncs()
				proxy.ProxyForMiddleware(func(in struct {
					x gone.Logger `gone:"xxx"`
				}) {
				})

			})

			t.Run("Bind Context Error", func(t *testing.T) {
				bindErr := errors.New("bind error")

				injector.EXPECT().StartBindFuncs()
				injector.EXPECT().BindFuncs().Return(func(ctx *gin.OriginContent, obj any, T reflect.Type) (reflect.Value, error) {
					return reflect.Value{}, bindErr
				})

				responser.EXPECT().Failed(gomock.Any(), gomock.Any()).Do(func(ctx any, err error) {
					assert.Equal(t, bindErr, err)
				})

				arr := proxy.ProxyForMiddleware(func(in struct {
					x gone.Logger `gone:"*"`
				}) {
				})
				arr[0](&gin.OriginContent{})
			})
		})
}
