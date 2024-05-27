package gone_grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/gone-io/gone"
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
	gone.Logger `gone:"gone-logger"`

	port int    `gone:"config,server.grpc.port,default=9090"`
	host string `gone:"config,server.grpc.host,default=0.0.0.0"`

	grpcServer *grpc.Server
	listener   net.Listener

	grpcServices []Service `gone:"*"`
	gone.Tracer  `gone:"gone-tracer"`

	address string

	createListener func(*server) error
}

func NewServer() gone.Goner {
	return &server{
		createListener: createListener,
	}
}

func (s *server) initListener(cemetery gone.Cemetery) error {
	tomb := cemetery.GetTomById(gone.IdGoneCMux)
	if tomb != nil {
		cMux := tomb.GetGoner().(gone.CMuxServer)
		//s.listener = cMux.Match(
		//	cmux.HTTP2HeaderField("content-type", "application/grpc"),
		//	cmux.HTTP1HeaderField("content-type", "application/grpc"),
		//)

		s.listener = cMux.MatchWithWriters(
			cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
		)

		s.address = cMux.GetAddress()
		return nil
	}
	s.address = fmt.Sprintf("%s:%d", s.host, s.port)
	return s.createListener(s)
}

func (s *server) Start(cemetery gone.Cemetery) error {
	if len(s.grpcServices) == 0 {
		return errors.New("no gRPC service found, gRPC server will not start")
	}

	err := s.initListener(cemetery)
	if err != nil {
		return err
	}

	s.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.traceInterceptor,
			s.recoveryInterceptor,
		),
	)

	for _, grpcService := range s.grpcServices {
		s.Infof("Register gRPC service %v", reflect.ValueOf(grpcService).Type().String())
		grpcService.RegisterGrpcServer(s.grpcServer)
	}

	s.Infof("gRPC server now listen at %s", s.address)
	s.Go(s.server)

	return nil
}

func (s *server) server() {
	if err := s.grpcServer.Serve(s.listener); err != nil {
		s.Errorf("failed to serve: %v", err)
	}
}

func (s *server) Stop(gone.Cemetery) error {
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

	s.SetTraceId(traceId, func() {
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
	defer s.Recover()
	return handler(ctx, req)
}
