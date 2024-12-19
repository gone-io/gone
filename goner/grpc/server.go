package gone_grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/gone-io/gone"
	gonecmux "github.com/gone-io/gone/goner/cmux"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"reflect"
)

const XTraceId = "x-trace-id"

func createListener(s *server) (err error) {
	s.listener, err = net.Listen("tcp", s.address)
	return
}

type server struct {
	gone.Flag
	logger       gone.Logger      `gone:"*"`
	tracer       gone.Tracer      `gone:"*"`
	grpcServices []Service        `gone:"*"`
	keeper       gone.GonerKeeper `gone:"*"`

	port int    `gone:"config,server.grpc.port,default=9090"`
	host string `gone:"config,server.grpc.host,default=0.0.0.0"`

	grpcServer     *grpc.Server
	listener       net.Listener
	address        string
	createListener func(*server) error
}

func (s *server) GonerName() string {
	return "gone-grpc-server"
}

func (s *server) initListener() error {
	goner := s.keeper.GetGonerByName(gonecmux.Name)
	if goner != nil {
		if muxServer, ok := goner.(gone.CMuxServer); ok {
			s.listener = muxServer.MatchWithWriters(
				cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
			)
			s.address = muxServer.GetAddress()
			return nil
		}
	}
	s.address = fmt.Sprintf("%s:%d", s.host, s.port)
	return s.createListener(s)
}
func (s *server) Init() error {
	if len(s.grpcServices) == 0 {
		return errors.New("no gRPC service found, gRPC server will not start")
	}
	err := s.initListener()
	if err != nil {
		return gone.ToError(err)
	}

	return nil
}

func (s *server) register() {
	s.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.traceInterceptor,
			s.recoveryInterceptor,
		),
	)
	for _, grpcService := range s.grpcServices {
		s.logger.Infof("Register gRPC service %v", reflect.ValueOf(grpcService).Type().String())
		grpcService.RegisterGrpcServer(s.grpcServer)
	}
}

func (s *server) Start() error {
	s.register()
	s.tracer.Go(s.server)
	return nil
}

func (s *server) server() {
	s.logger.Infof("gRPC server now listen at %s", s.address)
	if err := s.grpcServer.Serve(s.listener); err != nil {
		s.logger.Errorf("failed to serve: %v", err)
	}
}

func (s *server) Stop() error {
	s.grpcServer.Stop()
	return nil
}

func (s *server) traceInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	var traceId string
	traceIdV := metadata.ValueFromIncomingContext(ctx, XTraceId)
	if len(traceIdV) > 0 {
		traceId = traceIdV[0]
	}

	s.tracer.SetTraceId(traceId, func() {
		resp, err = handler(ctx, req)
	})
	return
}

func (s *server) recoveryInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	defer s.tracer.RecoverAndSetError(&err)
	return handler(ctx, req)
}
