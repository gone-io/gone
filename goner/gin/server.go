package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/cmux"
	"github.com/gone-io/gone/goner/logrus"
	Cmux "github.com/soheilhy/cmux"
	"net/http"
)

func NewGinServer() (gone.Angel, gone.GonerId) {
	return &server{}, gone.IdGoneGin
}

type server struct {
	gone.Flag
	httpServer    *http.Server
	logrus.Logger `gone:"gone-logger"`
	http.Handler  `gone:"gone-gin-router"`

	net         cmux.Server  `gone:"gone-cumx"`
	mode        string       `gone:"config,server.mode,default=release"`
	controllers []Controller `gone:"*"`
}

func (s *server) Start(gone.Cemetery) error {
	//设置模式
	gin.SetMode(s.mode)

	s.mount()

	l := s.net.Match(Cmux.HTTP1Fast())

	s.httpServer = &http.Server{
		Handler: s,
	}

	s.Infof("Server Listen At %s", s.net.GetAddress())
	go func() {
		if err := s.httpServer.Serve(l); err != nil && err != http.ErrServerClosed {
			s.Errorf("http server error: %v", err)
			panic(err)
		}
	}()
	return nil
}

func (s *server) Stop(gone.Cemetery) error {
	if nil == s.httpServer {
		return nil
	}
	return s.httpServer.Close()
}

func (s *server) Serve() Close {
	_ = s.Start(nil)
	return func() {
		_ = s.Stop(nil)
	}
}

// 挂载路由
func (s *server) mount() {
	if len(s.controllers) == 0 {
		s.Warnf("There is no controller working")
	}

	for _, c := range s.controllers {
		err := c.Mount()
		if err != nil {
			s.Errorf("controller mount err:%v", err)
			panic(err)
		}
	}
}
