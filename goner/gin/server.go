package gin

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/cmux"
	Cmux "github.com/soheilhy/cmux"
	"net"
	"net/http"
	"sync"
)

func NewGinServer() (gone.Goner, gone.GonerId) {
	s := server{}
	return &s, gone.IdGoneGin
}

type server struct {
	gone.Flag
	httpServer   *http.Server
	gone.Logger  `gone:"gone-logger"`
	http.Handler `gone:"gone-gin-router"`
	gone.Tracer  `gone:"gone-tracer"`

	net         cmux.Server  `gone:"gone-cumx"`
	mode        string       `gone:"config,server.mode,default=release"`
	controllers []Controller `gone:"*"`

	l        net.Listener
	stopFlag bool
	lock     sync.Mutex
}

func (s *server) Start(gone.Cemetery) error {
	err := s.mount()
	if err != nil {
		return err
	}
	s.setServer()

	s.Infof("Server Listen At %s", s.net.GetAddress())
	s.Go(s.serve)
	return nil
}
func (s *server) setServer() {
	s.stopFlag = false
	gin.SetMode(s.mode) //设置模式
	s.l = s.net.Match(Cmux.HTTP1Fast(http.MethodPatch))
	s.httpServer = &http.Server{
		Handler: s,
	}
}

func (s *server) serve() {
	if err := s.httpServer.Serve(s.l); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

	err = s.l.Close()
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
