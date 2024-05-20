package cmux

import (
	"github.com/soheilhy/cmux"
	"net"
)

//go:generate sh -c "mockgen -package=cmux net Listener,Conn > net_Listener_mock_test.go"

// Server cumx 服务，用于复用同一端口监听多种协议，参考文档：https://pkg.go.dev/github.com/soheilhy/cmux
type Server interface {
	Match(matcher ...cmux.Matcher) net.Listener
	GetAddress() string
}
