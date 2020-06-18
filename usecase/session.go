package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/domain/repository"
	"github.com/camphor-/relaym-server/domain/service"
	"github.com/camphor-/relaym-server/domain/spotify"
	"github.com/camphor-/relaym-server/log"
)

var syncCheckOffset = 5 * time.Second

// SessionUseCase はセッションに関するユースケースです。
type SessionUseCase struct {
	tm          *entity.SyncCheckTimerManager
	sessionRepo repository.Session
	userRepo    repository.User
	playerCli   spotify.Player
	trackCli    spotify.TrackClient
	pusher      event.Pusher
}

// NewSessionUseCase はSessionUseCaseのポインタを生成します。
func NewSessionUseCase(sessionRepo repository.Session, userRepo repository.User, playerCli spotify.Player, trackCli spotify.TrackClient, pusher event.Pusher) *SessionUseCase {
	return &SessionUseCase{
		tm:          entity.NewSyncCheckTimerManager(),
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		playerCli:   playerCli,
		trackCli:    trackCli,
		pusher:      pusher,
	}
}

// AddQueueTrack はセッションのqueueにTrackを追加します。
func (s *SessionUseCase) AddQueueTrack(ctx context.Context, sessionID string, trackURI string) error {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return fmt.Errorf("FindByID sessionID=%s: %w", sessionID, err)
	}

	err = s.sessionRepo.StoreQueueTrack(&entity.QueueTrackToStore{
		URI:       trackURI,
		SessionID: sessionID,
	})
	if err != nil {
		return fmt.Errorf("StoreQueueTrack URI=%s, sessionID=%s: %w", trackURI, sessionID, err)
	}

	if session.ShouldCallAddQueueAPINow() {
		err = s.playerCli.AddToQueue(ctx, trackURI, session.DeviceID)
		if err != nil {
			return fmt.Errorf("AddToQueue URI=%s, sessionID=%s: %w", trackURI, sessionID, err)
		}
	}
	s.pusher.Push(&event.PushMessage{
		SessionID: sessionID,
		Msg:       entity.EventAddTrack,
	})

	return nil
}

// CreateSession は与えられたセッション名のセッションを作成します。
func (s *SessionUseCase) CreateSession(sessionName string, creatorID string) (*entity.SessionWithUser, error) {
	creator, err := s.userRepo.FindByID(creatorID)
	if err != nil {
		return nil, fmt.Errorf("FindByID userID=%s: %w", creatorID, err)
	}

	newSession, err := entity.NewSession(sessionName, creatorID)
	if err != nil {
		return nil, fmt.Errorf("NewSession sessionName=%s: %w", sessionName, err)
	}

	err = s.sessionRepo.StoreSession(newSession)
	if err != nil {
		return nil, fmt.Errorf("StoreSession sessionName=%s: %w", sessionName, err)
	}
	return entity.NewSessionWithUser(newSession, creator), nil
}

// ChangePlaybackState は与えられたセッションの再生状態を操作します。
func (s *SessionUseCase) ChangePlaybackState(ctx context.Context, sessionID string, st entity.StateType) error {
	switch st {
	case entity.Play:
		if err := s.play(ctx, sessionID); err != nil {
			return fmt.Errorf("play sessionID=%s: %w", sessionID, err)
		}
	case entity.Pause:
		if err := s.pause(ctx, sessionID); err != nil {
			return fmt.Errorf("pause sessionID=%s: %w", sessionID, err)
		}
	}
	return nil
}

// Play はセッションのstateを STOP → PLAY に変更して曲の再生を始めます。
func (s *SessionUseCase) play(ctx context.Context, sessionID string) error {
	sess, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return fmt.Errorf("find session id=%s: %w", sessionID, err)
	}

	if len(sess.QueueTracks) == 0 {
		return fmt.Errorf("play: %w", entity.ErrQueueTrackNotFound)
	}

	if err := s.playerCli.SetRepeatMode(ctx, false, ""); err != nil {
		return fmt.Errorf("call set repeat off api: %w", err)
	}

	if err := s.playerCli.SetShuffleMode(ctx, false, ""); err != nil {
		return fmt.Errorf("call set repeat off api: %w", err)
	}

	// TODO : デバイスIDをどっかから読み込む
	if sess.IsResume(entity.Play) {
		if err := s.playerCli.Play(ctx, ""); err != nil {
			return fmt.Errorf("call play api: %w", err)
		}
	} else {
		if err := s.playerCli.PlayWithTracks(ctx, "", sess.TrackURIsOnAndAfterQueueHead()[:1]); err != nil {
			return fmt.Errorf("call play api with tracks %v: %w", sess.TrackURIsOnAndAfterQueueHead(), err)
		}
		for i := 1; i < len(sess.TrackURIsOnAndAfterQueueHead()); i++ {
			trackURI := sess.TrackURIsOnAndAfterQueueHead()[i]
			if err := s.playerCli.AddToQueue(ctx, trackURI, ""); err != nil {
				return fmt.Errorf("call add queue api trackURI=%s: %w", trackURI, err)
			}
		}
	}

	if err := sess.MoveToPlay(); err != nil {
		return fmt.Errorf("move to play id=%s: %w", sessionID, err)
	}

	if err := s.sessionRepo.Update(sess); err != nil {
		return fmt.Errorf("update session id=%s: %w", sessionID, err)
	}

	go s.startTrackEndTrigger(ctx, sessionID)

	s.pusher.Push(&event.PushMessage{
		SessionID: sessionID,
		Msg:       entity.EventPlay,
	})

	return nil
}

// Pause はセッションのstateをPLAY→PAUSEに変更して曲の再生を一時停止します。
func (s *SessionUseCase) pause(ctx context.Context, sessionID string) error {
	sess, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return fmt.Errorf("find session id=%s: %w", sessionID, err)
	}

	// TODO : デバイスIDをどっかから読み込む
	if err := s.playerCli.Pause(ctx, ""); err != nil && !errors.Is(err, entity.ErrActiveDeviceNotFound) {
		return fmt.Errorf("call pause api: %w", err)
	}

	s.tm.StopTimer(sessionID)

	if err := sess.MoveToPause(); err != nil {
		return fmt.Errorf("move to pause id=%s: %w", sessionID, err)
	}

	if err := s.sessionRepo.Update(sess); err != nil {
		return fmt.Errorf("update session id=%s: %w", sessionID, err)
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: sessionID,
		Msg:       entity.EventPause,
	})

	return nil
}

// CanConnectToPusher はイベントをクライアントにプッシュするためのコネクションを貼れるかどうかチェックします。
func (s *SessionUseCase) CanConnectToPusher(ctx context.Context, sessionID string) error {
	sess, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return fmt.Errorf("find session id=%s: %w", sessionID, err)
	}

	// セッションが再生中なのに同期チェックがされていなかったら始める
	// サーバ再起動でタイマーがなくなると、イベントが正しくクライアントに送られなくなるのでこのタイミングで復旧させる。
	if _, ok := s.tm.GetTimer(sessionID); !ok && sess.IsPlaying() {
		fmt.Printf("session timer not found: create timer: sessionID=%s\n", sessionID)
		go s.startTrackEndTrigger(ctx, sessionID)
	}

	return nil
}

// startTrackEndTrigger は曲の終了やストップを検知してそれぞれの処理を実行します。 goroutineで実行されることを想定しています。
func (s *SessionUseCase) startTrackEndTrigger(ctx context.Context, sessionID string) {
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "start track end trigger", "sessionID": sessionID})

	time.Sleep(5 * time.Second) // 曲の再生が始まるのを待つ
	playingInfo, err := s.playerCli.CurrentlyPlaying(ctx)
	if err != nil {
		logger.Errorj(map[string]interface{}{
			"message":   "startTrackEndTrigger: failed to get currently playing info",
			"sessionID": sessionID,
			"error":     err,
		})
		return
	}
	remainDuration := playingInfo.Remain()

	logger.Infoj(map[string]interface{}{
		"message": "start timer", "sessionID": sessionID, "remainDuration": remainDuration.String(),
	})

	triggerAfterTrackEnd := s.tm.CreateTimer(sessionID, remainDuration+syncCheckOffset)

	for {
		select {
		case <-triggerAfterTrackEnd.StopCh():
			logger.Infoj(map[string]interface{}{"message": "stop timer", "sessionID": sessionID})
			return
		case <-triggerAfterTrackEnd.ExpireCh():
			logger.Debugj(map[string]interface{}{"message": "trigger expired", "sessionID": sessionID})
			timer, nextTrack, err := s.handleTrackEnd(ctx, sessionID)
			if err != nil {
				if errors.Is(err, entity.ErrSessionPlayingDifferentTrack) {
					logger.Infoj(map[string]interface{}{"message": "handleTrackEnd detects interrupt", "sessionID": sessionID, "error": err.Error()})
					return
				}
				logger.Errorj(map[string]interface{}{"message": "handleTrackEnd with error", "sessionID": sessionID, "error": err.Error()})
				return
			}
			if !nextTrack {
				logger.Infoj(map[string]interface{}{"message": "no next track", "sessionID": sessionID})
				return
			}
			triggerAfterTrackEnd = timer
		}
	}
}

// handleTrackEnd はある一曲の再生が終わったときの処理を行います。
func (s *SessionUseCase) handleTrackEnd(ctx context.Context, sessionID string) (triggerAfterTrackEnd *entity.SyncCheckTimer, nextTrack bool, returnErr error) {
	logger := log.New()

	sess, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return nil, false, fmt.Errorf("find session id=%s: %v", sessionID, err)
	}

	defer func() {
		if err := s.sessionRepo.Update(sess); err != nil {
			if returnErr != nil {
				returnErr = fmt.Errorf("update session id=%s: %v: %w", sess.ID, err, returnErr)
			} else {
				returnErr = fmt.Errorf("update session id=%s: %w", sess.ID, err)
			}
		}
	}()

	if err := sess.GoNextTrack(); err != nil && errors.Is(err, entity.ErrSessionAllTracksFinished) {
		s.handleAllTrackFinish(sess)
		return nil, false, nil
	}

	logger.Debugj(map[string]interface{}{"message": "next track", "sessionID": sessionID, "queueHead": sess.QueueHead})

	playingInfo, err := s.playerCli.CurrentlyPlaying(ctx)
	if err != nil {
		if errors.Is(err, entity.ErrActiveDeviceNotFound) {
			if interErr := s.handleInterrupt(sess); interErr != nil {
				returnErr = fmt.Errorf("handle interrupt: %w", interErr)
				return nil, false, returnErr
			}
			returnErr = err
			return nil, false, returnErr
		}
		returnErr = fmt.Errorf("get currently playing info id=%s: %v", sessionID, err)
		return nil, false, returnErr
	}

	if err := sess.IsPlayingCorrectTrack(playingInfo); err != nil {
		if interErr := s.handleInterrupt(sess); interErr != nil {
			returnErr = fmt.Errorf("check whether playing correct track: handle interrupt: %v: %w", interErr, err)
			return nil, false, returnErr
		}
		returnErr = fmt.Errorf("check whether playing correct track: %w", err)
		return nil, false, returnErr
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: sessionID,
		Msg:       entity.NewEventNextTrack(sess.QueueHead),
	})
	triggerAfterTrackEnd = s.tm.CreateTimer(sessionID, playingInfo.Remain()+syncCheckOffset)

	logger.Infoj(map[string]interface{}{
		"message": "restart timer", "sessionID": sessionID, "remainDuration": playingInfo.Remain().String(),
	})

	return triggerAfterTrackEnd, true, nil

}

// handleAllTrackFinish はキューの全ての曲の再生が終わったときの処理を行います。
func (s *SessionUseCase) handleAllTrackFinish(sess *entity.Session) {
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "all track finished", "sessionID": sess.ID})
	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.EventStop,
	})
}

// handleInterrupt はSpotifyとの同期が取れていないときの処理を行います。
func (s *SessionUseCase) handleInterrupt(sess *entity.Session) error {
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "interrupt detected", "sessionID": sess.ID})

	if err := sess.MoveToStop(); err != nil {
		// 必ずPlayされているのでこのエラーになることはないはず
		logger.Errorj(map[string]interface{}{"message": "failed to move to stop", "sessionID": sess.ID, "error": err.Error()})
		return fmt.Errorf("move to stop: id=%s: %v", sess.ID, err)
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.EventInterrupt,
	})
	s.tm.DeleteTimer(sess.ID)
	return nil
}

// SetDevice は指定されたidのセッションの作成者と再生する端末を紐付けて再生するデバイスを指定します。
func (s *SessionUseCase) SetDevice(ctx context.Context, sessionID string, deviceID string) error {
	userID, ok := service.GetUserIDFromContext(ctx)
	if !ok {
		return errors.New("get user id from context")
	}

	sess, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return fmt.Errorf("find session id=%s: %w", sessionID, err)
	}

	if !sess.IsCreator(userID) {
		return fmt.Errorf("userID=%s creatorID=%s: %w", userID, sess.CreatorID, entity.ErrUserIsNotSessionCreator)
	}

	sess.DeviceID = deviceID
	if err := s.sessionRepo.Update(sess); err != nil {
		return fmt.Errorf("update device id: device_id=%s session_id=%s: %w", deviceID, sess.ID, err)
	}

	return nil
}

// GetSession は指定されたidからsessionの情報を返します
func (s *SessionUseCase) GetSession(ctx context.Context, sessionID string) (*entity.SessionWithUser, []*entity.Track, *entity.CurrentPlayingInfo, error) {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("FindByID sessionID=%s: %w", sessionID, err)
	}

	creator, err := s.userRepo.FindByID(session.CreatorID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("FindByID userID=%s: %w", session.CreatorID, err)
	}

	trackURIs := make([]string, len(session.QueueTracks))
	for i, queueTrack := range session.QueueTracks {
		trackURIs[i] = queueTrack.URI
	}

	tracks, err := s.trackCli.GetTracksFromURI(ctx, trackURIs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get tracks: track_uris=%s: %w", trackURIs, err)
	}

	cpi, err := s.playerCli.CurrentlyPlaying(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("CurrentlyPlaying: %w", err)
	}

	return entity.NewSessionWithUser(session, creator), tracks, cpi, nil
}
