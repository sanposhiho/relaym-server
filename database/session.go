package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/oauth2"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/repository"

	"github.com/go-gorp/gorp/v3"
	"github.com/go-sql-driver/mysql"
)

var _ repository.Session = &SessionRepository{}

var errorNumDuplicateEntry uint16 = 1062

// SessionRepository は repository.SessionRepository を満たす構造体です
type SessionRepository struct {
	dbMap *gorp.DbMap
}

// NewSessionRepository はSessionRepositoryのポインタを生成する関数です
func NewSessionRepository(dbMap *gorp.DbMap) *SessionRepository {
	dbMap.AddTableWithName(sessionDTO{}, "sessions").SetKeys(false, "ID")
	dbMap.AddTableWithName(queueTrackDTO{}, "queue_tracks")
	return &SessionRepository{dbMap: dbMap}
}

// FindByID は指定されたIDを持つsessionをDBから取得します
func (r *SessionRepository) FindByID(id string) (*entity.Session, error) {
	var dto sessionDTO
	if err := r.dbMap.SelectOne(&dto, "SELECT id, name, creator_id, queue_head, state_type, device_id, expired_at FROM sessions WHERE id = ?", id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("select session: %w", entity.ErrSessionNotFound)
		}
		return nil, fmt.Errorf("select session: %w", err)
	}

	queueTracks, errOnGetQueue := r.getQueueTracksBySessionID(id)
	if errOnGetQueue != nil {
		return nil, fmt.Errorf("get queue tracks: %w", errOnGetQueue)
	}

	stateType, err := entity.NewStateType(dto.StateType)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", entity.ErrInvalidStateType)
	}

	return &entity.Session{
		ID:          dto.ID,
		Name:        dto.Name,
		CreatorID:   dto.CreatorID,
		DeviceID:    dto.DeviceID,
		StateType:   stateType,
		QueueHead:   dto.QueueHead,
		QueueTracks: queueTracks,
		ExpiredAt:   dto.ExpiredAt,
	}, nil
}

// FindCreatorTokenBySessionID はSessionIDからCreatorのTokenを取得します
func (r *SessionRepository) FindCreatorTokenBySessionID(sessionID string) (*oauth2.Token, string, error) {
	var dto spotifyAuthDTO

	if err := r.dbMap.SelectOne(&dto, "SELECT sa.access_token, sa.refresh_token, sa.expiry, sessions.creator_id AS user_id FROM sessions INNER JOIN spotify_auth AS sa ON sa.user_id = sessions.creator_id WHERE sessions.id = ?", sessionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", fmt.Errorf("select session: %w", entity.ErrSessionNotFound)
		}
		return nil, "", fmt.Errorf("select session: %w", err)
	}

	return &oauth2.Token{
		AccessToken:  dto.AccessToken,
		TokenType:    "Bearer",
		RefreshToken: dto.RefreshToken,
		Expiry:       dto.Expiry,
	}, dto.UserID, nil
}

// StoreSession はSessionをDBに挿入します。
func (r *SessionRepository) StoreSession(session *entity.Session) error {
	threeDaysAfter := time.Now().AddDate(0, 0, 3).UTC()
	dto := &sessionDTO{
		ID:        session.ID,
		Name:      session.Name,
		CreatorID: session.CreatorID,
		QueueHead: session.QueueHead,
		StateType: session.StateType.String(),
		DeviceID:  session.DeviceID,
		ExpiredAt: threeDaysAfter,
	}

	if err := r.dbMap.Insert(dto); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == errorNumDuplicateEntry {
			return fmt.Errorf("insert session: %w", entity.ErrSessionAlreadyExisted)
		}
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// Update はセッションの情報を更新します。
func (r *SessionRepository) Update(session *entity.Session) error {
	dto := &sessionDTO{
		ID:        session.ID,
		Name:      session.Name,
		CreatorID: session.CreatorID,
		QueueHead: session.QueueHead,
		StateType: session.StateType.String(),
		DeviceID:  session.DeviceID,
		ExpiredAt: session.ExpiredAt,
	}

	if _, err := r.dbMap.Update(dto); err != nil {
		return fmt.Errorf("update session: %w", err)
	}
	return nil
}

// UpdateWithExpiredAt はセッションの情報を更新し、同時にExpiredAtを更新します。
func (r *SessionRepository) UpdateWithExpiredAt(session *entity.Session, newExpiredAt time.Time) error {
	dto := &sessionDTO{
		ID:        session.ID,
		Name:      session.Name,
		CreatorID: session.CreatorID,
		QueueHead: session.QueueHead,
		StateType: session.StateType.String(),
		DeviceID:  session.DeviceID,
		ExpiredAt: newExpiredAt,
	}

	if _, err := r.dbMap.Update(dto); err != nil {
		return fmt.Errorf("update session: %w", err)
	}
	return nil
}

// StoreQueueTrack はQueueTrackをDBに挿入します。
func (r *SessionRepository) StoreQueueTrack(queueTrack *entity.QueueTrackToStore) error {
	if _, err := r.dbMap.Exec("INSERT INTO queue_tracks(`index`, uri, session_id) SELECT COALESCE(MAX(`index`),-1)+1, ?, ? from queue_tracks as qt WHERE session_id = ?;", queueTrack.URI, queueTrack.SessionID, queueTrack.SessionID); err != nil {
		return fmt.Errorf("insert queue_tracks: %w", err)
	}
	return nil
}

// ArchiveSessionsForBatch は以下の条件に当てはまるSessionのstateをArchivedに変更します
//// - 作成から3日以上が経過している。もしくはArchiveが解除されてから3日以上が経過している
func (r *SessionRepository) ArchiveSessionsForBatch() error {
	currentDateTime := time.Now()
	if _, err := r.dbMap.Exec("UPDATE sessions SET state_type = 'ARCHIVED' WHERE state_type != 'ARCHIVED' AND expired_at < ?;", currentDateTime); err != nil {
		return fmt.Errorf("update session state_type to ARCHIVED: %w", err)
	}
	return nil
}

func (r *SessionRepository) getQueueTracksBySessionID(id string) ([]*entity.QueueTrack, error) {
	var dto []queueTrackDTO
	if _, err := r.dbMap.Select(&dto, "SELECT * FROM queue_tracks WHERE session_id = ? ORDER BY `index` ASC", id); err != nil {
		return nil, fmt.Errorf("select queue_tracks: %w", err)
	}
	return r.toQueueTracks(dto), nil
}

func (r *SessionRepository) toQueueTracks(resultQueueTracks []queueTrackDTO) []*entity.QueueTrack {
	queueTracks := make([]*entity.QueueTrack, len(resultQueueTracks))

	for i, rs := range resultQueueTracks {
		queueTracks[i] = &entity.QueueTrack{
			Index:     rs.Index,
			URI:       rs.URI,
			SessionID: rs.SessionID,
		}
	}

	return queueTracks
}

type sessionDTO struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CreatorID string    `db:"creator_id"`
	QueueHead int       `db:"queue_head"`
	StateType string    `db:"state_type"`
	DeviceID  string    `db:"device_id"`
	ExpiredAt time.Time `db:"expired_at"`
}

type queueTrackDTO struct {
	Index     int    `db:"index"`
	URI       string `db:"uri"`
	SessionID string `db:"session_id"`
}
