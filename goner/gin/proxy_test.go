package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_proxy_proxyOne(t *testing.T) {
	type fields struct {
		Flag      gone.Flag
		Logger    gone.Logger
		cemetery  gone.Cemetery
		responser Responser
		tracer    gone.Tracer
		inject    func(logger gone.Logger, cemetery gone.Cemetery, responser Responser, x HandlerFunc, context *gin.Context) (results []any)
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockResponser := NewMockResponser(controller)
	mockResponser.EXPECT().ProcessResults(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockResponser.EXPECT().ProcessResults(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockResponser.EXPECT().ProcessResults(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	gone.Prepare().AfterStart(func(in struct {
		cemetery gone.Cemetery `gone:"gone-cemetery"`
	}) {

		type args struct {
			x    HandlerFunc
			last bool
		}
		tests := []struct {
			name     string
			fields   fields
			args     args
			wantFunc func(want gin.HandlerFunc) bool
		}{
			{
				name: "func(*Context) (any, error)",
				fields: fields{
					responser: mockResponser,
				},
				args: args{
					x:    func(ctx *Context) (any, error) { return nil, nil },
					last: false,
				},
				wantFunc: func(want gin.HandlerFunc) bool {
					want(&gin.Context{})
					return true
				},
			},
			{
				name: "func(*Context) error",
				fields: fields{
					responser: mockResponser,
				},
				args: args{
					x:    func(*Context) error { return nil },
					last: false,
				},
				wantFunc: func(want gin.HandlerFunc) bool {
					want(&gin.Context{})
					return true
				},
			},
			{
				name: "func(*Context)",
				fields: fields{
					responser: mockResponser,
				},
				args: args{
					x:    func(*Context) {},
					last: false,
				},
				wantFunc: func(want gin.HandlerFunc) bool {
					want(&gin.Context{})
					return true
				},
			},
			{
				name: "other",
				fields: fields{
					responser: mockResponser,
					cemetery:  in.cemetery,
				},
				args: args{
					x:    func(in struct{}) {},
					last: false,
				},
				wantFunc: func(want gin.HandlerFunc) bool {
					want(&gin.Context{})
					return true
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				p := &proxy{
					Flag:      tt.fields.Flag,
					Logger:    tt.fields.Logger,
					cemetery:  tt.fields.cemetery,
					responser: tt.fields.responser,
					tracer:    tt.fields.tracer,
				}
				one := p.proxyOne(tt.args.x, tt.args.last)

				assert.Truef(t, tt.wantFunc(one), "proxyOne(%v, %v)", tt.args.x, tt.args.last)
			})
		}
	}).Run()
}

func Test_proxy_ProxyForMiddleware(t *testing.T) {
	ginProxy, _, _ := NewGinProxy()
	p := ginProxy.(HandleProxyToGin)
	funcs := p.ProxyForMiddleware(func(ctx *gin.Context) {}, func() {})
	assert.Equal(t, 2, len(funcs))
}

func Test_proxy_Proxy(t *testing.T) {
	ginProxy, _, _ := NewGinProxy()
	p := ginProxy.(HandleProxyToGin)
	funcs := p.Proxy(func(ctx *gin.Context) {}, func() {})
	assert.Equal(t, 2, len(funcs))
}
