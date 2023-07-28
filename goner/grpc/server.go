package gone_grpc

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"google.golang.org/grpc"
	"net"
	"reflect"
)

type Service interface {
	RegisterGrpcServer(server *grpc.Server)
}

type Server struct {
	gone.Goner
	logrus.Logger `gone:"gone-logger"`

	port int `gone:"config,grpc.server.port"`

	grpcServices []Service `gone:"*"`
	grpcServer   *grpc.Server
	options      []grpc.ServerOption
}

func NewServer(opt ...grpc.ServerOption) *Server {
	return &Server{
		options: opt,
	}
}

func (g Server) Start(gone.Cemetery) error {
	if len(g.grpcServices) == 0 {
		g.Warnf("No grpc service found, grpc server will not start")
		return nil
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	g.grpcServer = grpc.NewServer(g.options...)
	for _, grpcService := range g.grpcServices {
		g.Infof("Register grpc service %v", reflect.ValueOf(grpcService).Type().String())
		grpcService.RegisterGrpcServer(g.grpcServer)
	}

	go func() {
		g.Infof("Grpc server now listen at %d", g.port)
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
