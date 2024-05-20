package cmux

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/gone-io/gone"
	Cmux "github.com/soheilhy/cmux"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func (l *server) Go(fn func()) {
	go fn()
}
func (l *server) Errorf(format string, args ...any) {}
func (l *server) Warnf(format string, args ...any)  {}

func Test_cumx(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	s := server{
		listen: func(network, address string) (net.Listener, error) {
			return nil, errors.New("not support")
		},
	}
	err := s.AfterRevive()
	assert.Error(t, err)
}

func Test_cumx2(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	s := server{
		listen: func(network, address string) (net.Listener, error) {
			return nil, errors.New("not support")
		},
	}
	listener := NewMockListener(controller)
	conn := NewMockConn(controller)
	conn.EXPECT().Close().Return(nil).AnyTimes()

	listener.EXPECT().Accept().Return(conn, nil).AnyTimes()

	s.listen = func(network, address string) (net.Listener, error) {
		return listener, nil
	}
	err := s.AfterRevive()
	cemeteryForTest := gone.NewBuryMockCemeteryForTest()

	assert.Nil(t, err)
	err = s.Start(cemeteryForTest)
	assert.Nil(t, err)

	_ = s.GetAddress()

	err = s.Stop(cemeteryForTest)
	assert.Nil(t, err)
}

func Test_server_Match_And_Error(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	listener := NewMockListener(controller)
	conn := NewMockConn(controller)
	conn.EXPECT().Close().Return(nil).AnyTimes()

	s := server{
		listen: func(network, address string) (net.Listener, error) {
			return listener, nil
		},
	}
	err := s.AfterRevive()
	assert.Nil(t, err)

	match := s.Match(Cmux.HTTP1Fast())

	assert.NotNil(t, match)

	listener.EXPECT().Accept().Return(nil, errors.New("error")).AnyTimes()
	cemeteryForTest := gone.NewBuryMockCemeteryForTest()
	err = s.Start(cemeteryForTest)
	assert.Error(t, err)
}

func Test_server_processStartError(t *testing.T) {
	s := server{}
	s.processStartError(errors.New("error"))
	s.stopFlag = true
	s.processStartError(errors.New("error"))
}
