package main

import (
	"context"
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner"
	"google.golang.org/grpc"
	"grpc/proto"
	"log"
)

type helloClient struct {
	gone.Flag
	proto.HelloClient

	host string `gone:"config,server.host"`
	port string `gone:"config,server.port"`
}

func (c *helloClient) Address() string {
	return fmt.Sprintf("%s:%s", c.host, c.port)
}

func (c *helloClient) Stub(conn *grpc.ClientConn) {
	c.HelloClient = proto.NewHelloClient(conn)
}

func main() {
	gone.Prepare(func(cemetery gone.Cemetery) error {
		_ = goner.BasePriest(cemetery)
		_ = goner.GrpcClientPriest(cemetery)

		cemetery.Bury(&helloClient{})
		return nil
	}).AfterStart(func(in struct {
		hello *helloClient `gone:"*"`
	}) {
		say, err := in.hello.Say(context.Background(), &proto.SayRequest{Name: "gone"})
		if err != nil {
			log.Printf("er:%v", err)
			return
		}
		log.Printf("say result: %s", say.Message)
	}).Run()
}
