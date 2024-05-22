// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go

// Package schedule is a generated GoMock package.
package schedule

import (
	goneMock "github.com/gone-io/gone"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockScheduler is a mock of Scheduler interface.
type MockScheduler struct {
	goneMock.Flag
	ctrl     *gomock.Controller
	recorder *MockSchedulerMockRecorder
}

// MockSchedulerMockRecorder is the mock recorder for MockScheduler.
type MockSchedulerMockRecorder struct {
	mock *MockScheduler
}

// NewMockScheduler creates a new mock instance.
func NewMockScheduler(ctrl *gomock.Controller) *MockScheduler {
	mock := &MockScheduler{ctrl: ctrl}
	mock.recorder = &MockSchedulerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockScheduler) EXPECT() *MockSchedulerMockRecorder {
	return m.recorder
}

// Cron mocks base method.
func (m *MockScheduler) Cron(run RunFuncOnceAt) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Cron", run)
}

// Cron indicates an expected call of Cron.
func (mr *MockSchedulerMockRecorder) Cron(run interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Cron", reflect.TypeOf((*MockScheduler)(nil).Cron), run)
}

// MockSchedule is a mock of Schedule interface.
type MockSchedule struct {
	goneMock.Flag
	ctrl     *gomock.Controller
	recorder *MockScheduleMockRecorder
}

// MockScheduleMockRecorder is the mock recorder for MockSchedule.
type MockScheduleMockRecorder struct {
	mock *MockSchedule
}

// NewMockSchedule creates a new mock instance.
func NewMockSchedule(ctrl *gomock.Controller) *MockSchedule {
	mock := &MockSchedule{ctrl: ctrl}
	mock.recorder = &MockScheduleMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSchedule) EXPECT() *MockScheduleMockRecorder {
	return m.recorder
}

// Start mocks base method.
func (m *MockSchedule) Start() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockScheduleMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockSchedule)(nil).Start))
}

// Stop mocks base method.
func (m *MockSchedule) Stop() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop")
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockScheduleMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockSchedule)(nil).Stop))
}
