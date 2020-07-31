package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/domain/repository"
	"github.com/camphor-/relaym-server/domain/spotify"
	"github.com/camphor-/relaym-server/log"
)

var waitTimeBeforeHandleTrackEnd = 7 * time.Second

type SessionTimerUseCase struct {
	tm          *entity.SyncCheckTimerManager
	sessionRepo repository.Session
	playerCli   spotify.Player
	pusher      event.Pusher
}

func NewSessionTimerUseCase(sessionRepo repository.Session, playerCli spotify.Player, pusher event.Pusher, tm *entity.SyncCheckTimerManager) *SessionTimerUseCase {
	return &SessionTimerUseCase{tm: tm, sessionRepo: sessionRepo, playerCli: playerCli, pusher: pusher}
}

// startTrackEndTrigger は曲の終了やストップを検知してそれぞれの処理を実行します。 goroutineで実行されることを想定しています。
func (s *SessionTimerUseCase) startTrackEndTrigger(ctx context.Context, sessionID string) {
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "start track end trigger", "sessionID": sessionID})

	// 曲の再生を待つ
	waitPlayTimer := time.NewTimer(5 * time.Second)

	triggerAfterTrackEnd := s.tm.CreateTimer(sessionID)
	for {
		select {
		case <-waitPlayTimer.C:
			playingInfo, err := s.playerCli.CurrentlyPlaying(ctx)
			if err != nil {
				logger.Errorj(map[string]interface{}{
					"message":   "startTrackEndTrigger: failed to get currently playing info",
					"sessionID": sessionID,
					"error":     err.Error(),
				})
				return
			}

			// ぴったしのタイマーをセットすると、Spotifyでは次の曲の再生が始まってるのにRelaym側では次の曲に進んでおらず、
			// INTERRUPTになってしまう
			remainDuration := playingInfo.Remain() - 2*time.Second

			logger.Infoj(map[string]interface{}{
				"message": "start timer", "sessionID": sessionID, "remainDuration": remainDuration.String(),
			})

			triggerAfterTrackEnd.SetTimer(remainDuration)

		case <-triggerAfterTrackEnd.StopCh():
			if !waitPlayTimer.Stop() {
				<-waitPlayTimer.C
			}
			logger.Infoj(map[string]interface{}{"message": "stop timer", "sessionID": sessionID})
			return

		case <-triggerAfterTrackEnd.NextCh():
			logger.Debugj(map[string]interface{}{"message": "call to move next track", "sessionID": sessionID})
			waitPlayTimer.Stop()
			timer, nextTrack, err := s.handleSkipTrack(ctx, sessionID)
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

		case <-triggerAfterTrackEnd.ExpireCh():
			logger.Debugj(map[string]interface{}{"message": "trigger expired", "sessionID": sessionID})
			waitPlayTimer.Stop()
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

// handleSkipTrack はある一曲のSkipが呼び出された時の処理を行います。
func (s *SessionTimerUseCase) handleSkipTrack(ctx context.Context, sessionID string) (*entity.SyncCheckTimer, bool, error) {
	s.tm.DeleteTimer(sessionID)

	triggerAfterTrackEndResponse, err := s.sessionRepo.DoInTx(ctx, s.handleSkipTrackTx(sessionID))
	if v, ok := triggerAfterTrackEndResponse.(*handleTrackEndResponse); ok {
		// これはトランザクションが失敗してRollbackしたとき
		if err != nil {
			return nil, false, fmt.Errorf("handle track end in transaction: %w", err)
		}
		return v.triggerAfterTrackEnd, v.nextTrack, v.err
	}
	// これはトランザクションが失敗してRollbackしたとき
	if err != nil {
		return nil, false, fmt.Errorf("handle track end in transaction: %w", err)
	}
	return nil, false, nil
}

// handleTrackEnd はある一曲の再生が終わったときの処理を行います。
func (s *SessionTimerUseCase) handleTrackEnd(ctx context.Context, sessionID string) (*entity.SyncCheckTimer, bool, error) {
	s.tm.DeleteTimer(sessionID)
	time.Sleep(waitTimeBeforeHandleTrackEnd)

	triggerAfterTrackEndResponse, err := s.sessionRepo.DoInTx(ctx, s.handleTrackEndTx(sessionID))
	if v, ok := triggerAfterTrackEndResponse.(*handleTrackEndResponse); ok {
		// これはトランザクションが失敗してRollbackしたとき
		if err != nil {
			return nil, false, fmt.Errorf("handle track end in transaction: %w", err)
		}
		return v.triggerAfterTrackEnd, v.nextTrack, v.err
	}
	// これはトランザクションが失敗してRollbackしたとき
	if err != nil {
		return nil, false, fmt.Errorf("handle track end in transaction: %w", err)
	}
	return nil, false, nil
}

// handleTrackEndTx はINTERRUPTになってerrorを帰す場合もトランザクションをコミットして欲しいので、
// アプリケーションエラーはhandleTrackEndResponseのフィールドで返すようにしてerrorの返り値はnilにしている
func (s *SessionTimerUseCase) handleTrackEndTx(sessionID string) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (_ interface{}, returnErr error) {
		sess, err := s.sessionRepo.FindByIDForUpdate(ctx, sessionID)
		if err != nil {
			return &handleTrackEndResponse{triggerAfterTrackEnd: nil, nextTrack: false}, fmt.Errorf("find session id=%s: %v", sessionID, err)
		}

		return s.UpdateForHandleNextTrackTx(ctx, sess)
	}
}

// handleSkipTrackTx はINTERRUPTになってerrorを帰す場合もトランザクションをコミットして欲しいので、
// アプリケーションエラーはhandleTrackEndResponseのフィールドで返すようにしてerrorの返り値はnilにしている
func (s *SessionTimerUseCase) handleSkipTrackTx(sessionID string) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (_ interface{}, returnErr error) {
		sess, err := s.sessionRepo.FindByIDForUpdate(ctx, sessionID)
		if err != nil {
			return &handleTrackEndResponse{triggerAfterTrackEnd: nil, nextTrack: false}, fmt.Errorf("find session id=%s: %v", sessionID, err)
		}

		// ロックが取られてないタイミングでskipの反映待ちをするとdifferent trackになるので、
		// このタイミングでskipの反映待ちをしている
		time.Sleep(300 * time.Millisecond)

		return s.UpdateForHandleNextTrackTx(ctx, sess)
	}
}

func (s *SessionTimerUseCase) UpdateForHandleNextTrackTx(ctx context.Context, sess *entity.Session) (_ interface{}, returnErr error) {
	logger := log.New()

	defer func() {
		if err := s.sessionRepo.Update(ctx, sess); err != nil {
			if returnErr != nil {
				returnErr = fmt.Errorf("update session id=%s: %v: %w", sess.ID, err, returnErr)
			} else {
				returnErr = fmt.Errorf("update session id=%s: %w", sess.ID, err)
			}
		}
	}()

	// 曲の再生中にArchivedになった場合
	if sess.StateType == entity.Archived {

		s.pusher.Push(&event.PushMessage{
			SessionID: sess.ID,
			Msg:       entity.EventArchived,
		})

		return &handleTrackEndResponse{
			triggerAfterTrackEnd: nil,
			nextTrack:            false,
			err:                  nil,
		}, nil
	}

	if err := sess.GoNextTrack(); err != nil && errors.Is(err, entity.ErrSessionAllTracksFinished) {
		s.handleAllTrackFinish(sess)
		return &handleTrackEndResponse{
			triggerAfterTrackEnd: nil,
			nextTrack:            false,
			err:                  nil,
		}, nil
	}

	track := sess.TrackURIShouldBeAddedWhenHandleTrackEnd()
	if track != "" {
		if err := s.playerCli.Enqueue(ctx, track, sess.DeviceID); err != nil {
			return &handleTrackEndResponse{
				triggerAfterTrackEnd: nil,
				nextTrack:            false,
				err:                  fmt.Errorf("call add queue api trackURI=%s: %w", track, err),
			}, nil
		}
	}

	logger.Debugj(map[string]interface{}{"message": "next track", "sessionID": sess.ID, "queueHead": sess.QueueHead})

	playingInfo, err := s.playerCli.CurrentlyPlaying(ctx)
	if err != nil {
		if errors.Is(err, entity.ErrActiveDeviceNotFound) {
			s.handleInterrupt(sess)
			return &handleTrackEndResponse{
				triggerAfterTrackEnd: nil,
				nextTrack:            false,
				err:                  err,
			}, nil
		}
		return &handleTrackEndResponse{triggerAfterTrackEnd: nil, nextTrack: false, err: fmt.Errorf("get currently playing info id=%s: %v", sess.ID, err)}, nil
	}

	if err := sess.IsPlayingCorrectTrack(playingInfo); err != nil {
		s.handleInterrupt(sess)
		return &handleTrackEndResponse{triggerAfterTrackEnd: nil, nextTrack: false, err: fmt.Errorf("check whether playing correct track: %w", err)}, nil
	}

	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.NewEventNextTrack(sess.QueueHead),
	})

	triggerAfterTrackEnd := s.tm.CreateTimer(sess.ID)

	triggerAfterTrackEnd.SetTimer(playingInfo.Remain())

	logger.Infoj(map[string]interface{}{
		"message": "restart timer", "sessionID": sess.ID, "remainDuration": playingInfo.Remain().String(),
	})

	return &handleTrackEndResponse{triggerAfterTrackEnd: triggerAfterTrackEnd, nextTrack: true, err: nil}, nil
}

// handleAllTrackFinish はキューの全ての曲の再生が終わったときの処理を行います。
func (s *SessionTimerUseCase) handleAllTrackFinish(sess *entity.Session) {
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "all track finished", "sessionID": sess.ID})
	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.EventStop,
	})
}

// handleInterrupt はSpotifyとの同期が取れていないときの処理を行います。
func (s *SessionTimerUseCase) handleInterrupt(sess *entity.Session) {
	logger := log.New()
	logger.Debugj(map[string]interface{}{"message": "interrupt detected", "sessionID": sess.ID})

	sess.MoveToStop()

	s.pusher.Push(&event.PushMessage{
		SessionID: sess.ID,
		Msg:       entity.EventInterrupt,
	})
}

func (s *SessionTimerUseCase) existsTimer(sessionID string) bool {
	_, exists := s.tm.GetTimer(sessionID)
	return exists
}

func (s *SessionTimerUseCase) stopTimer(sessionID string) {
	s.tm.StopTimer(sessionID)
}

func (s *SessionTimerUseCase) deleteTimer(sessionID string) {
	s.tm.DeleteTimer(sessionID)
}

type handleTrackEndResponse struct {
	triggerAfterTrackEnd *entity.SyncCheckTimer
	nextTrack            bool
	err                  error
}
