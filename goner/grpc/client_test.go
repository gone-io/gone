package gone_grpc

import (
	"context"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"testing"
)

func (s *clientRegister) Infof(format string, args ...any) {}
func TestClientRegister_traceInterceptor(t *testing.T) {
	gone.Prepare(config.Priest, tracer.Priest, logrus.Priest).AfterStart(func(in struct {
		tracer tracer.Tracer `gone:"gone-tracer"`
	}) {

		var req, reply any

		register := clientRegister{
			tracer: in.tracer,
		}
		tracer.SetTraceId("xxxx", func() {
			err := register.traceInterceptor(
				context.Background(),
				"test",
				req, reply,
				nil,
				func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
					md, b := metadata.FromOutgoingContext(ctx)
					assert.True(t, b)
					list := md[XTraceId]

					assert.Equal(t, "xxxx", list[0])
					return nil
				},
			)
			assert.Nil(t, err)
		})
	}).Run()
}

func Test_clientRegister_register(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	client := NewMockClient(controller)
	client.EXPECT().Address().Return(":8080").AnyTimes()
	client.EXPECT().Stub(gomock.Any())

	register := clientRegister{
		connections: make(map[string]*grpc.ClientConn),
		clients:     []Client{client},
	}

	err := register.Start(nil)
	assert.Nil(t, err)
}

func Test_clientRegister_Stop(t *testing.T) {
	register := clientRegister{
		connections: make(map[string]*grpc.ClientConn),
	}

	conn, err2 := grpc.NewClient(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.Nil(t, err2)
	register.connections[":8080"] = conn

	err := register.Stop(nil)
	assert.Nil(t, err)
}
