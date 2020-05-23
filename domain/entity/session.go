package entity

import (
	"fmt"

	"github.com/google/uuid"
)

// Session はセッションを表します。
type Session struct {
	ID          string
	Name        string
	CreatorID   string
	StateType   StateType
	QueueHead   int
	QueueTracks []*QueueTrack
}

type SessionWithUser struct {
	*Session
	Creator *User
}

// NewSession はSessionのポインタを生成する関数です。
func NewSession(name string, creatorID string) (*Session, error) {
	return &Session{
		ID:          uuid.New().String(),
		Name:        name,
		CreatorID:   creatorID,
		QueueHead:   0,
		StateType:   Stop,
		QueueTracks: nil,
	}, nil
}

// NewSessionWithUser はSession(ポインタ)からSessionWithUser(ポインタ)を生成します
func NewSessionWithUser(session *Session, creator *User) *SessionWithUser {
	return &SessionWithUser{
		Session: &Session{
			ID:          session.ID,
			Name:        session.Name,
			CreatorID:   creator.ID,
			StateType:   session.StateType,
			QueueHead:   session.QueueHead,
			QueueTracks: session.QueueTracks,
		},
		Creator: creator,
	}
}

// MoveToPlay はセッションのStateTypeをPlayに状態遷移します。
func (s *Session) MoveToPlay() error {
	s.StateType = Play
	return nil

}

// MoveToPause はセッションのStateTypeをPauseに状態遷移します。
func (s *Session) MoveToPause() error {
	if s.StateType == Play || s.StateType == Pause {
		s.StateType = Pause
		return nil
	}
	return fmt.Errorf("state type from %s to Pause: %w", s.StateType, ErrChangeSessionStateNotPermit)
}

type StateType string

const (
	Play  StateType = "PLAY"
	Pause StateType = "PAUSE"
	Stop  StateType = "STOP"
)

var stateTypes = []StateType{Play, Pause, Stop}

// NewStateType はstringから対応するStateTypeを生成します。
func NewStateType(stateType string) (StateType, error) {
	for _, st := range stateTypes {
		if st.String() == stateType {
			return st, nil
		}
	}
	return "", fmt.Errorf("stateType = %s:%w", stateType, ErrInvalidStateType)
}

// String はfmt.Stringerを満たすメソッドです。
func (st StateType) String() string {
	return string(st)
}
