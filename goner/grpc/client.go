package gone_grpc

import (
	"context"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client interface {
	Address() string
	Stub(conn *grpc.ClientConn)
}

type ClientRegister struct {
	gone.Goner
	connections   map[string]*grpc.ClientConn
	clients       []Client `gone:"*"`
	tracer.Tracer `gone:"gone-tracer"`
}

//go:gone
func NewRegister() gone.Goner {
	return &ClientRegister{connections: make(map[string]*grpc.ClientConn)}
}

func (s *ClientRegister) TraceInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, XTraceId, tracer.GetTraceId())
		return invoker(ctx, method, req, reply, cc)
	}
}

func (s *ClientRegister) register(client Client) error {
	conn, ok := s.connections[client.Address()]
	if !ok {
		c, err := grpc.Dial(client.Address(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(s.TraceInterceptor()),
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

func (s *ClientRegister) Start(gone.Cemetery) error {
	for _, c := range s.clients {
		if err := s.register(c); err != nil {
			return err
		}
	}

	return nil
}

func (s ClientRegister) Stop(gone.Cemetery) error {
	for _, conn := range s.connections {
		conn.Close()
	}
	return nil
}
