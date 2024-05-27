package main

import (
	"context"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner"
	"github.com/gone-io/gone/goner/cmux"
	"google.golang.org/grpc"
	"grpc/proto"
	"log"
)

type server struct {
	gone.Flag
	proto.UnimplementedHelloServer
}

// Say implements Hello.Say
func (s *server) Say(ctx context.Context, in *proto.SayRequest) (*proto.SayResponse, error) {
	log.Printf("Received: %v", in.GetName())
	return &proto.SayResponse{Message: "Hello " + in.GetName()}, nil
}

func (s *server) RegisterGrpcServer(server *grpc.Server) {
	proto.RegisterHelloServer(server, s)
}

func main() {
	gone.Prepare(func(cemetery gone.Cemetery) error {
		_ = goner.BasePriest(cemetery)
		_ = cmux.Priest(cemetery)
		_ = goner.GrpcServerPriest(cemetery)

		cemetery.Bury(&server{})
		return nil
	}).Serve()
}
