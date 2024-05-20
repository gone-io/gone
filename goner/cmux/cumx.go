package cmux

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/soheilhy/cmux"
	"net"
	"sync"
	"time"
)

func NewServer() (gone.Angel, gone.GonerId) {
	s := server{}
	s.listen = net.Listen
	return &s, gone.IdGoneCumx
}

type server struct {
	gone.Flag
	once        sync.Once
	c           cmux.CMux
	gone.Logger `gone:"gone-logger"`
	gone.Tracer `gone:"gone-tracer"`

	stopFlag bool
	lock     sync.Mutex

	Network string `gone:"config,server.network,default=tcp"`
	Address string `gone:"config,server.address"`
	Host    string `gone:"config,server.host"`
	Port    int    `gone:"config,server.port,default=8080"`

	listen func(network, address string) (net.Listener, error)
}

func (l *server) AfterRevive() gone.AfterReviveError {
	var err error
	if l.c == nil {
		l.once.Do(func() {
			if l.Address == "" {
				l.Address = fmt.Sprintf("%s:%d", l.Host, l.Port)
			}
			var listener net.Listener
			listener, err = l.listen(l.Network, l.Address)
			l.c = cmux.New(listener)
		})
	}
	return err
}

func (l *server) Match(matcher ...cmux.Matcher) net.Listener {
	return l.c.Match(matcher...)
}

func (l *server) GetAddress() string {
	return l.Address
}

func (l *server) Start(gone.Cemetery) error {
	l.stopFlag = false
	var err error
	l.Go(func() {
		err = l.c.Serve()
		l.processStartError(err)
	})
	<-time.After(10 * time.Millisecond)
	return err
}
func (l *server) processStartError(err error) {
	if err != nil {
		l.lock.Lock()
		if l.stopFlag {
			l.Errorf("cumx Serve() err:%v", err)
		} else {
			l.Warnf("cumx Serve() err:%v", err)
		}
		l.lock.Unlock()
	}
}

func (l *server) Stop(gone.Cemetery) error {
	l.Warnf("cumx server stopping!!")
	l.lock.Lock()
	l.stopFlag = true
	l.lock.Unlock()
	l.c.Close()
	return nil
}
