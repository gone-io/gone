package cmux

import (
	"github.com/gone-io/gone"
	"github.com/soheilhy/cmux"
	"net"
)

type Server interface {
	gone.Angel
	Match(matcher ...cmux.Matcher) net.Listener
	GetAddress() string
}
