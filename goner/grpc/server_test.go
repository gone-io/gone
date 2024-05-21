package gone_grpc

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"testing"
)

func (s *server) Errorf(format string, args ...any) {}
func (s *server) Warnf(format string, args ...any)  {}
func (s *server) Infof(format string, args ...any)  {}
func (s *server) Go(fn func())                      {}

//func Test_createListener(t *testing.T) {
//	err := createListener(&server{})
//	assert.Nil(t, err)
//}

func Test_server_initListener(t *testing.T) {
	t.Run("use cMuxServer", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		cMuxServer := NewCmuxServer(controller)
		listener := NewMockListener(controller)
		cMuxServer.EXPECT().Match(gomock.Any()).Return(listener)
		cMuxServer.EXPECT().GetAddress().Return("")

		cemetery := gone.NewBuryMockCemeteryForTest()
		cemetery.Bury(cMuxServer, gone.IdGoneCMux)

		s := server{}
		err := s.initListener(cemetery)
		assert.Nil(t, err)
	})

	t.Run("use tcpListener", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()
		listener := NewMockListener(controller)

		s := server{
			createListener: func(s *server) error {
				s.listener = listener
				return nil
			},
		}
		cemetery := gone.NewBuryMockCemeteryForTest()
		err := s.initListener(cemetery)
		assert.Nil(t, err)
	})

	t.Run("use tcpListener error", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()
		listener := NewMockListener(controller)

		s := server{
			createListener: func(s *server) error {
				s.listener = listener
				return errors.New("error")
			},
		}
		cemetery := gone.NewBuryMockCemeteryForTest()
		err := s.initListener(cemetery)
		assert.Error(t, err)
	})
}

func Test_server_Start(t *testing.T) {
	t.Run("no gRPC service found, gRPC server will not start", func(t *testing.T) {
		cemetery := gone.NewBuryMockCemeteryForTest()
		s := server{}
		err := s.Start(cemetery)
		assert.Error(t, err)
	})

	t.Run("initListener error", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		service := NewMockService(controller)
		cemetery := gone.NewBuryMockCemeteryForTest()
		s := server{
			grpcServices: []Service{service},
			createListener: func(s *server) error {
				return errors.New("error")
			},
		}
		err := s.Start(cemetery)
		assert.Error(t, err)
	})

	t.Run("suc", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		service := NewMockService(controller)
		service.EXPECT().RegisterGrpcServer(gomock.Any())

		listener := NewMockListener(controller)

		cemetery := gone.NewBuryMockCemeteryForTest()
		s := server{
			grpcServices: []Service{service},
			createListener: func(s *server) error {
				s.listener = listener
				return nil
			},
		}
		err := s.Start(cemetery)
		assert.Nil(t, err)
	})
}

type addr struct{}

func (a *addr) Network() string {
	return "tcp"
}
func (a *addr) String() string {
	return ":8080"
}

func Test_server_server(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	listener := NewMockListener(controller)
	listener.EXPECT().Addr().Return(&addr{}).AnyTimes()
	listener.EXPECT().Accept().Return(nil, errors.New("error"))
	listener.EXPECT().Close().Return(nil)

	s := server{
		grpcServer: grpc.NewServer(),
		listener:   listener,
	}
	s.server()
}

func Test_server_Stop(t *testing.T) {
	s := server{
		grpcServer: grpc.NewServer(),
	}
	err := s.Stop(nil)
	assert.Nil(t, err)
}

func Test_server_traceInterceptor(t *testing.T) {
	ctx := context.Background()
	traceId := "trace"

	ctx = metadata.NewIncomingContext(ctx, metadata.MD{
		XTraceId: []string{traceId},
	})

	gone.Prepare(tracer.Priest, logrus.Priest, config.Priest).AfterStart(func(in struct {
		tracer tracer.Tracer `gone:"gone-tracer"`
	}) {
		s := server{
			Tracer: in.tracer,
		}

		var req any
		_, err := s.traceInterceptor(ctx, req, nil, func(ctx context.Context, req any) (any, error) {
			id := in.tracer.GetTraceId()
			assert.Equal(t, traceId, id)
			return nil, nil
		})
		assert.Nil(t, err)

	}).Run()
}

func Test_server_recoveryInterceptor(t *testing.T) {
	gone.Prepare(tracer.Priest, logrus.Priest, config.Priest).AfterStart(func(in struct {
		tracer tracer.Tracer `gone:"gone-tracer"`
	}) {
		s := server{
			Tracer: in.tracer,
		}
		_, err := s.recoveryInterceptor(context.Background(), nil, nil, func(ctx context.Context, req any) (any, error) {
			panic(errors.New("error"))
		})
		assert.Nil(t, err)
	})
}
