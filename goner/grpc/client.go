package gone_grpc

import (
	"context"
	"github.com/gone-io/gone"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"reflect"
)

type clientRegister struct {
	gone.Flag
	gone.Logger `gone:"gone-logger"`
	connections map[string]*grpc.ClientConn
	clients     []Client    `gone:"*"`
	tracer      gone.Tracer `gone:"gone-tracer"`
}

//go:gone
func NewRegister() gone.Goner {
	return &clientRegister{connections: make(map[string]*grpc.ClientConn)}
}

func (s *clientRegister) traceInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	_ ...grpc.CallOption,
) error {
	ctx = metadata.AppendToOutgoingContext(ctx, XTraceId, s.tracer.GetTraceId())
	return invoker(ctx, method, req, reply, cc)
}

func (s *clientRegister) register(client Client) error {
	conn, ok := s.connections[client.Address()]
	if !ok {
		c, err := grpc.Dial(
			client.Address(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithChainUnaryInterceptor(s.traceInterceptor),
		)
		if err != nil {
			return err
		}

		s.connections[client.Address()] = c
		conn = c
	}

	client.Stub(conn)
	return nil
}

func (s *clientRegister) Start(gone.Cemetery) error {
	for _, c := range s.clients {
		s.Infof("register gRPC client %v on address %v\n", reflect.ValueOf(c).Type().String(), c.Address())
		if err := s.register(c); err != nil {
			return err
		}
	}

	return nil
}

func (s *clientRegister) Stop(gone.Cemetery) error {
	for _, conn := range s.connections {
		err := conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
