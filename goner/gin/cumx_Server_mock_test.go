// Code generated by MockGen. DO NOT EDIT.
// Source: ../cmux/interface.go

// Package gin is a generated GoMock package.
package gin

import (
	goneMock "github.com/gone-io/gone"
	net "net"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	cmux "github.com/soheilhy/cmux"
)

// CmuxServer is a mock of Server interface.
type CmuxServer struct {
	goneMock.Flag
	ctrl     *gomock.Controller
	recorder *CmuxServerMockRecorder
}

// CmuxServerMockRecorder is the mock recorder for CmuxServer.
type CmuxServerMockRecorder struct {
	mock *CmuxServer
}

// NewCmuxServer creates a new mock instance.
func NewCmuxServer(ctrl *gomock.Controller) *CmuxServer {
	mock := &CmuxServer{ctrl: ctrl}
	mock.recorder = &CmuxServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *CmuxServer) EXPECT() *CmuxServerMockRecorder {
	return m.recorder
}

// GetAddress mocks base method.
func (m *CmuxServer) GetAddress() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAddress")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetAddress indicates an expected call of GetAddress.
func (mr *CmuxServerMockRecorder) GetAddress() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAddress", reflect.TypeOf((*CmuxServer)(nil).GetAddress))
}

// Match mocks base method.
func (m *CmuxServer) Match(matcher ...cmux.Matcher) net.Listener {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range matcher {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Match", varargs...)
	ret0, _ := ret[0].(net.Listener)
	return ret0
}

// Match indicates an expected call of Match.
func (mr *CmuxServerMockRecorder) Match(matcher ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Match", reflect.TypeOf((*CmuxServer)(nil).Match), matcher...)
}