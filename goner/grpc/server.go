package gone_grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"reflect"
	"runtime/debug"
)

const XTraceId = "x-trace-id"

type Service interface {
	RegisterGrpcServer(server *grpc.Server)
}

type Server struct {
	gone.Goner
	logrus.Logger `gone:"gone-logger"`

	port int `gone:"config,grpc.server.port"`

	grpcServices  []Service `gone:"*"`
	grpcServer    *grpc.Server
	tracer.Tracer `gone:"gone-tracer"`
}

func NewServer() *Server {
	return &Server{}
}

func (g Server) Start(gone.Cemetery) error {
	if len(g.grpcServices) == 0 {
		g.Warnf("No gRPC service found, gRPC server will not start")
		return nil
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	g.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(g.TraceInterceptor()),
		grpc.UnaryInterceptor(g.RecoveryInterceptor()),
	)

	for _, grpcService := range g.grpcServices {
		g.Infof("Register gRPC service %v", reflect.ValueOf(grpcService).Type().String())
		grpcService.RegisterGrpcServer(g.grpcServer)
	}

	go func() {
		g.Infof("gRPC server now listen at %d", g.port)
		if err := g.grpcServer.Serve(lis); err != nil {
			g.Errorf("failed to serve: %v", err)
		}
	}()

	return nil
}

func (g Server) Stop(gone.Cemetery) error {
	g.grpcServer.Stop()
	return nil
}

func (g Server) TraceInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var traceId string
		traceIdV := metadata.ValueFromIncomingContext(ctx, XTraceId)
		if len(traceIdV) > 0 {
			traceId = traceIdV[0]
		}

		g.SetTraceId(traceId, func() {
			resp, err = handler(ctx, req)
		})

		return
	}
}

func (g Server) RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				g.Errorf("Panic occurred: %v, \n%v", r, debug.Stack())
				if er, ok := r.(error); ok {
					err = er
				}
				err = errors.New("gRPC panic occurred")
			}
		}()

		return handler(ctx, req)
	}
}
