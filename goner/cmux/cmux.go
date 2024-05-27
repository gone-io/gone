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
	return &s, gone.IdGoneCMux
}

type server struct {
	gone.Flag
	once        sync.Once
	cMux        cmux.CMux
	gone.Logger `gone:"gone-logger"`
	gone.Tracer `gone:"gone-tracer"`

	stopFlag bool
	lock     sync.Mutex

	network string `gone:"config,server.network,default=tcp"`
	address string `gone:"config,server.address"`
	host    string `gone:"config,server.host"`
	port    int    `gone:"config,server.port,default=8080"`

	listen func(network, address string) (net.Listener, error)
}

func (s *server) AfterRevive() gone.AfterReviveError {
	var err error
	if s.cMux == nil {
		s.once.Do(func() {
			if s.address == "" {
				s.address = fmt.Sprintf("%s:%d", s.host, s.port)
			}
			var listener net.Listener
			listener, err = s.listen(s.network, s.address)
			s.cMux = cmux.New(listener)
		})
	}
	return err
}

func (s *server) Match(matcher ...cmux.Matcher) net.Listener {
	return s.cMux.Match(matcher...)
}

func (s *server) MatchWithWriters(matcher ...cmux.MatchWriter) net.Listener {
	return s.cMux.MatchWithWriters(matcher...)
}

func (s *server) GetAddress() string {
	return s.address
}

func (s *server) Start(gone.Cemetery) error {
	s.stopFlag = false
	var err error
	var mutex sync.Mutex
	s.Go(func() {
		mutex.Lock()
		err = s.cMux.Serve()
		mutex.Unlock()
		s.processStartError(err)
	})
	<-time.After(10 * time.Millisecond)
	return err
}
func (s *server) processStartError(err error) {
	if err != nil {
		s.lock.Lock()
		if s.stopFlag {
			s.Errorf("cMux Serve() err:%v", err)
		} else {
			s.Warnf("cMux Serve() err:%v", err)
		}
		s.lock.Unlock()
	}
}

func (s *server) Stop(gone.Cemetery) error {
	s.Warnf("cMux server stopping!!")
	s.lock.Lock()
	s.stopFlag = true
	s.lock.Unlock()
	s.cMux.Close()
	return nil
}
