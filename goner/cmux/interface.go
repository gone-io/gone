package cmux

import (
	"github.com/gone-io/gone"
	"github.com/soheilhy/cmux"
	"net"
)

// Server cumx 服务，用于复用同一端口监听多种协议，参考文档：https://pkg.go.dev/github.com/soheilhy/cmux
type Server interface {
	gone.Angel
	Match(matcher ...cmux.Matcher) net.Listener
	GetAddress() string
}
