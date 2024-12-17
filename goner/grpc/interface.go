package gone_grpc

import "google.golang.org/grpc"

//go:generate sh -c "mockgen -package=gone_grpc net Listener > net_listener_mock_test.go"
//go:generate sh -c "mockgen -package=gone_grpc -source=../../interface_for_goner.go -mock_names=Server=CmuxServer | gone mock -o gone_mock_test.go"

//go:generate sh -c "mockgen -package=gone_grpc -self_package=github.com/gone-io/gone/goner/grpc -source=interface.go -destination=mock_test.go"

type Client interface {
	Address() string
	Stub(conn *grpc.ClientConn)
}

type Service interface {
	RegisterGrpcServer(server *grpc.Server)
}
