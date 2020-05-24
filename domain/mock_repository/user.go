// Code generated by MockGen. DO NOT EDIT.
// Source: user.go

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	entity "github.com/camphor-/relaym-server/domain/entity"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockUser is a mock of User interface
type MockUser struct {
	ctrl     *gomock.Controller
	recorder *MockUserMockRecorder
}

// MockUserMockRecorder is the mock recorder for MockUser
type MockUserMockRecorder struct {
	mock *MockUser
}

// NewMockUser creates a new mock instance
func NewMockUser(ctrl *gomock.Controller) *MockUser {
	mock := &MockUser{ctrl: ctrl}
	mock.recorder = &MockUserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockUser) EXPECT() *MockUserMockRecorder {
	return m.recorder
}

// FindByID mocks base method
func (m *MockUser) FindByID(id string) (*entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByID", id)
	ret0, _ := ret[0].(*entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByID indicates an expected call of FindByID
func (mr *MockUserMockRecorder) FindByID(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockUser)(nil).FindByID), id)
}

// FindBySpotifyUserID mocks base method
func (m *MockUser) FindBySpotifyUserID(spotifyUserID string) (*entity.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindBySpotifyUserID", spotifyUserID)
	ret0, _ := ret[0].(*entity.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindBySpotifyUserID indicates an expected call of FindBySpotifyUserID
func (mr *MockUserMockRecorder) FindBySpotifyUserID(spotifyUserID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindBySpotifyUserID", reflect.TypeOf((*MockUser)(nil).FindBySpotifyUserID), spotifyUserID)
}

// Store mocks base method
func (m *MockUser) Store(user *entity.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", user)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store
func (mr *MockUserMockRecorder) Store(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockUser)(nil).Store), user)
}

// Update mocks base method
func (m *MockUser) Update(user *entity.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", user)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockUserMockRecorder) Update(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockUser)(nil).Update), user)
}
