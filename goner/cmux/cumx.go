package cmux

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/soheilhy/cmux"
	"net"
	"sync"
)

type server struct {
	gone.Flag
	once          sync.Once
	c             cmux.CMux
	logrus.Logger `gone:"gone-logger"`

	Network string `gone:"config,server.network,default=tcp"`
	Address string `gone:"config,server.address"`
	Host    string `gone:"config,server.host"`
	Port    int    `gone:"config,server.port,default=8080"`
}

func (l *server) AfterRevive(gone.Cemetery, gone.Tomb) gone.ReviveAfterError {
	if l.c == nil {
		l.once.Do(func() {
			if l.Address == "" {
				l.Address = fmt.Sprintf("%s:%d", l.Host, l.Port)
			}
			listen, err := net.Listen(l.Network, l.Address)
			if err != nil {
				panic(err)
			}
			l.c = cmux.New(listen)
		})
	}
	return nil
}

func (l *server) Match(matcher ...cmux.Matcher) net.Listener {
	return l.c.Match(matcher...)
}

func (l *server) GetAddress() string {
	return l.Address
}

func (l *server) Start(gone.Cemetery) error {
	go func() {
		err := l.c.Serve()
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func (l *server) Stop(gone.Cemetery) error {
	l.c.Close()
	return nil
}
