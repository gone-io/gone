package gin

import (
	"errors"
	"fmt"
	"github.com/gone-io/gone"
	Cmux "github.com/soheilhy/cmux"
	"net"
	"net/http"
	"sync"
)

func NewGinServer() (gone.Goner, gone.GonerOption, gone.GonerOption) {
	s := server{
		createListener: createListener,
	}
	return &s, gone.IdGoneGin, gone.Order2
}

func createListener(s *server) (err error) {
	s.listener, err = net.Listen("tcp", s.address)
	return
}

type server struct {
	gone.Flag
	httpServer   *http.Server
	gone.Logger  `gone:"gone-logger"`
	http.Handler `gone:"gone-gin-router"`
	gone.Tracer  `gone:"gone-tracer"`

	controllers []Controller `gone:"*"`

	address  string
	stopFlag bool
	lock     sync.Mutex

	listener net.Listener
	port     int    `gone:"config,server.port=8080"`
	host     string `gone:"config,server.host,default=0.0.0.0"`

	createListener func(*server) error
}

func (s *server) Start(cemetery gone.Cemetery) error {
	err := s.mount()
	if err != nil {
		return err
	}
	err = s.initListener(cemetery)
	if err != nil {
		return err
	}

	s.stopFlag = false
	s.httpServer = &http.Server{
		Handler: s,
	}

	s.Infof("Server Listen At http://%s", s.address)
	s.Go(s.serve)
	return nil
}

func (s *server) initListener(cemetery gone.Cemetery) error {
	tomb := cemetery.GetTomById(gone.IdGoneCMux)
	if tomb != nil {
		cMux := tomb.GetGoner().(gone.CMuxServer)
		s.listener = cMux.Match(Cmux.HTTP1Fast(http.MethodPatch))
		s.address = cMux.GetAddress()
		return nil
	}
	s.address = fmt.Sprintf("%s:%d", s.host, s.port)
	return s.createListener(s)
}

func (s *server) serve() {
	if err := s.httpServer.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.processServeError(err)
	}
}

func (s *server) processServeError(err error) {
	s.lock.Lock()
	if !s.stopFlag {
		s.Errorf("http server error: %v", err)
		panic(err)
	} else {
		s.Warnf("http server error: %v", err)
	}
	s.lock.Unlock()
}

func (s *server) Stop(gone.Cemetery) (err error) {
	s.Warnf("gin server stopping!!")
	if nil == s.httpServer {
		return nil
	}
	s.lock.Lock()
	s.stopFlag = true
	s.lock.Unlock()

	err = s.listener.Close()
	if err != nil {
		s.Errorf("err:%v", err)
	}

	return err
}

// 挂载路由
func (s *server) mount() error {
	if len(s.controllers) == 0 {
		s.Warnf("There is no controller working")
	}

	for _, c := range s.controllers {
		err := c.Mount()
		if err != nil {
			s.Errorf("controller mount err:%v", err)
			return err
		}
	}
	return nil
}
