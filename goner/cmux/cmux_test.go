package cmux

import (
	"errors"
	"github.com/gone-io/gone"
	"github.com/soheilhy/cmux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net"
	"testing"
	"time"
)

func TestServer_Init(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	listener := NewMockListener(controller)

	s := &server{
		network: "tcp",
		address: "localhost:8080",
		listen: func(network, address string) (net.Listener, error) {
			return listener, nil
		},
	}

	err := s.Init()
	assert.NoError(t, err)
	assert.NotNil(t, s.cMux)
}

func TestServer_Start_Stop(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	listener := NewMockListener(controller)
	gone.Prepare(Load).Test(func(s *server) {
		s.listen = func(network, address string) (net.Listener, error) {
			return listener, nil
		}

		// Initialize the server
		err := s.Init()
		assert.NoError(t, err)

		// Start the server
		err = s.Start()
		assert.NoError(t, err)

		// Give some time for the server to start
		time.Sleep(50 * time.Millisecond)

		// Stop the server
		err = s.Stop()
		assert.NoError(t, err)

	})
}

func TestServer_Init_Error(t *testing.T) {
	s := &server{
		network: "tcp",
		address: "invalid_address", // 使用无效地址来触发错误
		listen: func(network, address string) (net.Listener, error) {
			return nil, errors.New("failed to listen")
		},
	}

	err := s.Init()
	assert.Error(t, err)
}

func Test_server_processStartError(t *testing.T) {
	gone.Prepare(Load).Test(func(s *server) {
		s.processStartError(errors.New("test error"))
		s.processStartError(nil)
		s.stopFlag = true
		s.processStartError(errors.New("test error"))
	})
}

func Test_server_Start_Stop(t *testing.T) {
	gone.Prepare(Load).Test(func(s *server) {
		httpL := s.Match(cmux.HTTP1Fast())
		assert.NotNil(t, httpL)
		grpcL := s.MatchWithWriters(
			cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
		)
		assert.NotNil(t, grpcL)
		address := s.GetAddress()
		assert.Equal(t, ":8080", address)
		assert.Equal(t, s.GonerName(), Name)
	})
}
