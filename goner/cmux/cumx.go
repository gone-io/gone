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
	stopFlag      bool
	lock          sync.Mutex

	Network string `gone:"config,server.network,default=tcp"`
	Address string `gone:"config,server.address"`
	Host    string `gone:"config,server.host"`
	Port    int    `gone:"config,server.port,default=8080"`
}

func (l *server) AfterRevive() gone.AfterReviveError {
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
	l.stopFlag = false
	go func() {
		err := l.c.Serve()
		if err != nil {
			l.lock.Lock()
			if l.stopFlag {
				l.Errorf("cumx Serve() err:%v", err)
			} else {
				l.Warnf("cumx Serve() err:%v", err)
			}
			l.lock.Unlock()
		}
	}()
	return nil
}

func (l *server) Stop(gone.Cemetery) error {
	l.Warnf("cumx server stopping!!")
	l.lock.Lock()
	l.stopFlag = true
	l.lock.Unlock()
	l.c.Close()
	return nil
}
