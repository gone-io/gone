package gin

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"testing"
)

func (s *server) Errorf(format string, args ...any) {}
func (s *server) Warnf(format string, args ...any)  {}
func (s *server) Infof(format string, args ...any)  {}
func (s *server) Go(func())                         {}

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

	s := server{
		controllers: []Controller{&testController{}},
	}
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

	listener := NewMockListener(controller)
	listener.EXPECT().Close().Return(errors.New("error"))

	s := server{
		controllers: []Controller{&testController{}},
		createListener: func(s *server) error {
			s.listener = listener
			return nil
		},
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
	listener := NewMockListener(controller)

	s := server{
		controllers: []Controller{&testController{}, &errController{}},
		createListener: func(s *server) error {
			s.listener = listener
			return nil
		},
	}
	cemetery := gone.NewBuryMockCemeteryForTest()

	err := s.Stop(cemetery)
	assert.Nil(t, err)

	err = s.Start(cemetery)
	assert.Error(t, err)

	s = server{
		createListener: func(s *server) error {
			s.listener = listener
			return nil
		},
	}

	err = s.Start(cemetery)
	assert.Nil(t, err)
}

//func Test_createListener(t *testing.T) {
//	err := createListener(&server{})
//	assert.Nil(t, err)
//}

func Test_server_initListener(t *testing.T) {
	t.Run("use cMuxServer", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()

		cMuxServer := NewCmuxServer(controller)
		listener := NewMockListener(controller)
		cMuxServer.EXPECT().Match(gomock.Any()).Return(listener)
		cMuxServer.EXPECT().GetAddress().Return("")

		cemetery := gone.NewBuryMockCemeteryForTest()
		cemetery.Bury(cMuxServer, gone.IdGoneCMux)

		s := server{}
		err := s.initListener(cemetery)
		assert.Nil(t, err)
	})

	t.Run("use tcpListener", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()
		listener := NewMockListener(controller)

		s := server{
			createListener: func(s *server) error {
				s.listener = listener
				return nil
			},
		}
		cemetery := gone.NewBuryMockCemeteryForTest()
		err := s.initListener(cemetery)
		assert.Nil(t, err)
	})

	t.Run("use tcpListener error", func(t *testing.T) {
		controller := gomock.NewController(t)
		defer controller.Finish()
		listener := NewMockListener(controller)

		s := server{
			createListener: func(s *server) error {
				s.listener = listener
				return errors.New("error")
			},
		}
		cemetery := gone.NewBuryMockCemeteryForTest()
		err := s.initListener(cemetery)
		assert.Error(t, err)
	})
}
