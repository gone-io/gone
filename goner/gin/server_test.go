package gin

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"testing"
)

func (r *server) Errorf(format string, args ...any) {}
func (r *server) Warnf(format string, args ...any)  {}
func (r *server) Infof(format string, args ...any)  {}
func (r *server) Go(func())                         {}

type testController struct {
}

func (t *testController) Mount() MountError {
	return nil
}

type errController struct {
}

func (t *errController) Mount() MountError {
	return errors.New("error")
}

func Test_server(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	cumxServer := NewCumxServer(controller)
	cumxServer.EXPECT().Match(gomock.Any()).Return(nil)

	s := server{
		net:         cumxServer,
		controllers: []Controller{&testController{}},
	}
	s.setServer()
	s.stopFlag = true
	s.processServeError(errors.New("error"))

	s.stopFlag = false
	func() {
		defer func() {
			err := recover()
			assert.Error(t, err.(error))
		}()
		s.processServeError(errors.New("error"))
	}()
}

func Test_server_start_stop(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	cumxServer := NewCumxServer(controller)
	listener := NewMockListener(controller)
	listener.EXPECT().Close().Return(errors.New("error"))

	cumxServer.EXPECT().GetAddress().Return(":8080")
	cumxServer.EXPECT().Match(gomock.Any()).Return(listener)
	s := server{
		net:         cumxServer,
		controllers: []Controller{&testController{}},
	}

	cemetery := gone.NewBuryMockCemeteryForTest()
	err := s.Start(cemetery)
	assert.Nil(t, err)

	err = s.Stop(cemetery)
	assert.Error(t, err)
}

func Test_server_error(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	cumxServer := NewCumxServer(controller)
	listener := NewMockListener(controller)
	cumxServer.EXPECT().Match(gomock.Any()).Return(listener)
	cumxServer.EXPECT().GetAddress().Return(":8080")

	s := server{
		net:         cumxServer,
		controllers: []Controller{&testController{}, &errController{}},
	}
	cemetery := gone.NewBuryMockCemeteryForTest()

	err := s.Stop(cemetery)
	assert.Nil(t, err)

	err = s.Start(cemetery)
	assert.Error(t, err)

	s = server{
		net: cumxServer,
	}

	err = s.Start(cemetery)
	assert.Nil(t, err)
}
