// Code generated by MockGen. DO NOT EDIT.
// Source: internal/models/users.go

// Package models is a generated GoMock package.
package models

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockUserStorage is a mock of UserStorage interface.
type MockUserStorage struct {
	ctrl     *gomock.Controller
	recorder *MockUserStorageMockRecorder
}

// MockUserStorageMockRecorder is the mock recorder for MockUserStorage.
type MockUserStorageMockRecorder struct {
	mock *MockUserStorage
}

// NewMockUserStorage creates a new mock instance.
func NewMockUserStorage(ctrl *gomock.Controller) *MockUserStorage {
	mock := &MockUserStorage{ctrl: ctrl}
	mock.recorder = &MockUserStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserStorage) EXPECT() *MockUserStorageMockRecorder {
	return m.recorder
}

// AddUser mocks base method.
func (m *MockUserStorage) AddUser(ctx context.Context, us *UserDTO) (*User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddUser", ctx, us)
	ret0, _ := ret[0].(*User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddUser indicates an expected call of AddUser.
func (mr *MockUserStorageMockRecorder) AddUser(ctx, us interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddUser", reflect.TypeOf((*MockUserStorage)(nil).AddUser), ctx, us)
}

// GetUser mocks base method.
func (m *MockUserStorage) GetUser(ctx context.Context, us *UserDTO) (*User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", ctx, us)
	ret0, _ := ret[0].(*User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockUserStorageMockRecorder) GetUser(ctx, us interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockUserStorage)(nil).GetUser), ctx, us)
}