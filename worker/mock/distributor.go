// Code generated by MockGen. DO NOT EDIT.
// Source: simplebank/worker (interfaces: TaskDistributor)
//
// Generated by this command:
//
//	mockgen -package mockwk -destination worker/mock/distributor.go simplebank/worker TaskDistributor
//

// Package mockwk is a generated GoMock package.
package mockwk

import (
	context "context"
	reflect "reflect"
	worker "simplebank/worker"

	asynq "github.com/hibiken/asynq"
	gomock "go.uber.org/mock/gomock"
)

// MockTaskDistributor is a mock of TaskDistributor interface.
type MockTaskDistributor struct {
	ctrl     *gomock.Controller
	recorder *MockTaskDistributorMockRecorder
	isgomock struct{}
}

// MockTaskDistributorMockRecorder is the mock recorder for MockTaskDistributor.
type MockTaskDistributorMockRecorder struct {
	mock *MockTaskDistributor
}

// NewMockTaskDistributor creates a new mock instance.
func NewMockTaskDistributor(ctrl *gomock.Controller) *MockTaskDistributor {
	mock := &MockTaskDistributor{ctrl: ctrl}
	mock.recorder = &MockTaskDistributorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTaskDistributor) EXPECT() *MockTaskDistributorMockRecorder {
	return m.recorder
}

// DistributeSendVerifyEmailTask mocks base method.
func (m *MockTaskDistributor) DistributeSendVerifyEmailTask(ctx context.Context, payload *worker.PayloadSendVerifyEmail, opts ...asynq.Option) error {
	m.ctrl.T.Helper()
	varargs := []any{ctx, payload}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DistributeSendVerifyEmailTask", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DistributeSendVerifyEmailTask indicates an expected call of DistributeSendVerifyEmailTask.
func (mr *MockTaskDistributorMockRecorder) DistributeSendVerifyEmailTask(ctx, payload any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, payload}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DistributeSendVerifyEmailTask", reflect.TypeOf((*MockTaskDistributor)(nil).DistributeSendVerifyEmailTask), varargs...)
}
