// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/argoproj/notifications-engine/pkg/api (interfaces: API)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	api "github.com/argoproj/notifications-engine/pkg/api"
	services "github.com/argoproj/notifications-engine/pkg/services"
	triggers "github.com/argoproj/notifications-engine/pkg/triggers"
	gomock "github.com/golang/mock/gomock"
)

// MockAPI is a mock of API interface.
type MockAPI struct {
	ctrl     *gomock.Controller
	recorder *MockAPIMockRecorder
}

// MockAPIMockRecorder is the mock recorder for MockAPI.
type MockAPIMockRecorder struct {
	mock *MockAPI
}

// NewMockAPI creates a new mock instance.
func NewMockAPI(ctrl *gomock.Controller) *MockAPI {
	mock := &MockAPI{ctrl: ctrl}
	mock.recorder = &MockAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAPI) EXPECT() *MockAPIMockRecorder {
	return m.recorder
}

// AddNotificationService mocks base method.
func (m *MockAPI) AddNotificationService(arg0 string, arg1 services.NotificationService) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddNotificationService", arg0, arg1)
}

// AddNotificationService indicates an expected call of AddNotificationService.
func (mr *MockAPIMockRecorder) AddNotificationService(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddNotificationService", reflect.TypeOf((*MockAPI)(nil).AddNotificationService), arg0, arg1)
}

// GetConfig mocks base method.
func (m *MockAPI) GetConfig() api.Config {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfig")
	ret0, _ := ret[0].(api.Config)
	return ret0
}

// GetConfig indicates an expected call of GetConfig.
func (mr *MockAPIMockRecorder) GetConfig() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfig", reflect.TypeOf((*MockAPI)(nil).GetConfig))
}

// GetNotificationServices mocks base method.
func (m *MockAPI) GetNotificationServices() map[string]services.NotificationService {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNotificationServices")
	ret0, _ := ret[0].(map[string]services.NotificationService)
	return ret0
}

// GetNotificationServices indicates an expected call of GetNotificationServices.
func (mr *MockAPIMockRecorder) GetNotificationServices() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNotificationServices", reflect.TypeOf((*MockAPI)(nil).GetNotificationServices))
}

// RunTrigger mocks base method.
func (m *MockAPI) RunTrigger(arg0 string, arg1 map[string]interface{}) ([]triggers.ConditionResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RunTrigger", arg0, arg1)
	ret0, _ := ret[0].([]triggers.ConditionResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RunTrigger indicates an expected call of RunTrigger.
func (mr *MockAPIMockRecorder) RunTrigger(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunTrigger", reflect.TypeOf((*MockAPI)(nil).RunTrigger), arg0, arg1)
}

// Send mocks base method.
func (m *MockAPI) Send(arg0 map[string]interface{}, arg1 []string, arg2 services.Destination) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockAPIMockRecorder) Send(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockAPI)(nil).Send), arg0, arg1, arg2)
}
