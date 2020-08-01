// Code generated by MockGen. DO NOT EDIT.
// Source: player.go

// Package mock_spotify is a generated GoMock package.
package mock_spotify

import (
	context "context"
	entity "github.com/camphor-/relaym-server/domain/entity"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// MockPlayer is a mock of Player interface
type MockPlayer struct {
	ctrl     *gomock.Controller
	recorder *MockPlayerMockRecorder
}

// MockPlayerMockRecorder is the mock recorder for MockPlayer
type MockPlayerMockRecorder struct {
	mock *MockPlayer
}

// NewMockPlayer creates a new mock instance
func NewMockPlayer(ctrl *gomock.Controller) *MockPlayer {
	mock := &MockPlayer{ctrl: ctrl}
	mock.recorder = &MockPlayerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPlayer) EXPECT() *MockPlayerMockRecorder {
	return m.recorder
}

// CurrentlyPlaying mocks base method
func (m *MockPlayer) CurrentlyPlaying(ctx context.Context) (*entity.CurrentPlayingInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CurrentlyPlaying", ctx)
	ret0, _ := ret[0].(*entity.CurrentPlayingInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CurrentlyPlaying indicates an expected call of CurrentlyPlaying
func (mr *MockPlayerMockRecorder) CurrentlyPlaying(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentlyPlaying", reflect.TypeOf((*MockPlayer)(nil).CurrentlyPlaying), ctx)
}

// Play mocks base method
func (m *MockPlayer) Play(ctx context.Context, deviceID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Play", ctx, deviceID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Play indicates an expected call of Play
func (mr *MockPlayerMockRecorder) Play(ctx, deviceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Play", reflect.TypeOf((*MockPlayer)(nil).Play), ctx, deviceID)
}

// PlayWithTracks mocks base method
func (m *MockPlayer) PlayWithTracks(ctx context.Context, deviceID string, trackURIs []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PlayWithTracks", ctx, deviceID, trackURIs)
	ret0, _ := ret[0].(error)
	return ret0
}

// PlayWithTracks indicates an expected call of PlayWithTracks
func (mr *MockPlayerMockRecorder) PlayWithTracks(ctx, deviceID, trackURIs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PlayWithTracks", reflect.TypeOf((*MockPlayer)(nil).PlayWithTracks), ctx, deviceID, trackURIs)
}

// PlayWithTracksAndPosition mocks base method
func (m *MockPlayer) PlayWithTracksAndPosition(ctx context.Context, deviceID string, trackURIs []string, position time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PlayWithTracksAndPosition", ctx, deviceID, trackURIs, position)
	ret0, _ := ret[0].(error)
	return ret0
}

// PlayWithTracksAndPosition indicates an expected call of PlayWithTracksAndPosition
func (mr *MockPlayerMockRecorder) PlayWithTracksAndPosition(ctx, deviceID, trackURIs, position interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PlayWithTracksAndPosition", reflect.TypeOf((*MockPlayer)(nil).PlayWithTracksAndPosition), ctx, deviceID, trackURIs, position)
}

// Pause mocks base method
func (m *MockPlayer) Pause(ctx context.Context, deviceID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Pause", ctx, deviceID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Pause indicates an expected call of Pause
func (mr *MockPlayerMockRecorder) Pause(ctx, deviceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pause", reflect.TypeOf((*MockPlayer)(nil).Pause), ctx, deviceID)
}

// Enqueue mocks base method
func (m *MockPlayer) Enqueue(ctx context.Context, trackURI, deviceID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Enqueue", ctx, trackURI, deviceID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Enqueue indicates an expected call of Enqueue
func (mr *MockPlayerMockRecorder) Enqueue(ctx, trackURI, deviceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enqueue", reflect.TypeOf((*MockPlayer)(nil).Enqueue), ctx, trackURI, deviceID)
}

// SetRepeatMode mocks base method
func (m *MockPlayer) SetRepeatMode(ctx context.Context, on bool, deviceID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetRepeatMode", ctx, on, deviceID)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetRepeatMode indicates an expected call of SetRepeatMode
func (mr *MockPlayerMockRecorder) SetRepeatMode(ctx, on, deviceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetRepeatMode", reflect.TypeOf((*MockPlayer)(nil).SetRepeatMode), ctx, on, deviceID)
}

// SetShuffleMode mocks base method
func (m *MockPlayer) SetShuffleMode(ctx context.Context, on bool, deviceID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetShuffleMode", ctx, on, deviceID)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetShuffleMode indicates an expected call of SetShuffleMode
func (mr *MockPlayerMockRecorder) SetShuffleMode(ctx, on, deviceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetShuffleMode", reflect.TypeOf((*MockPlayer)(nil).SetShuffleMode), ctx, on, deviceID)
}

// SkipAllTracks mocks base method
func (m *MockPlayer) SkipAllTracks(ctx context.Context, deviceID, trackURI string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SkipAllTracks", ctx, deviceID, trackURI)
	ret0, _ := ret[0].(error)
	return ret0
}

// SkipAllTracks indicates an expected call of SkipAllTracks
func (mr *MockPlayerMockRecorder) SkipAllTracks(ctx, deviceID, trackURI interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SkipAllTracks", reflect.TypeOf((*MockPlayer)(nil).SkipAllTracks), ctx, deviceID, trackURI)
}

// SkipCurrentTrack mocks base method
func (m *MockPlayer) SkipCurrentTrack(ctx context.Context, deviceID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SkipCurrentTrack", ctx, deviceID)
	ret0, _ := ret[0].(error)
	return ret0
}

// SkipCurrentTrack indicates an expected call of SkipCurrentTrack
func (mr *MockPlayerMockRecorder) SkipCurrentTrack(ctx, deviceID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SkipCurrentTrack", reflect.TypeOf((*MockPlayer)(nil).SkipCurrentTrack), ctx, deviceID)
}
