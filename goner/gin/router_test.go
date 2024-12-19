package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_router(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	handleProxyToGin := NewMockHandleProxyToGin(controller)
	handleProxyToGin.EXPECT().Proxy(gomock.Any()).Return(
		[]gin.HandlerFunc{
			func(ctx *gin.Context) {},
		},
	).AnyTimes()
	handleProxyToGin.EXPECT().ProxyForMiddleware(gomock.Any()).Return(
		[]gin.HandlerFunc{
			func(ctx *gin.Context) {},
		},
	).AnyTimes()

	r := router{
		HandleProxyToGin: handleProxyToGin,
	}

	err := r.Init()
	assert.Nil(t, err)

	fn := func() {}

	ginRouter := r.GetGinRouter()
	assert.Equal(t, ginRouter, r.Engine)
	r.Group("/api").Any("/test", fn)
	r.
		Group("/api/test2").
		Use(fn).
		GET("", fn).
		POST("", fn).
		DELETE("", fn).
		PATCH("", fn).
		PUT("", fn).
		OPTIONS("", fn).
		HEAD("", fn)

}
