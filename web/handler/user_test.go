package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/mock_repository"
	"github.com/camphor-/relaym-server/domain/mock_spotify"
	"github.com/camphor-/relaym-server/usecase"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/labstack/echo/v4"
)

func TestUserHandler_GetMe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		userID              string
		prepareMockUserRepo func(mock *mock_repository.MockUser)
		want                *userRes
		wantErr             bool
		wantCode            int
	}{
		{
			name:   "正しくユーザが取得できる",
			userID: "userID",
			prepareMockUserRepo: func(mock *mock_repository.MockUser) {
				mock.EXPECT().FindByID("userID").Return(&entity.User{
					ID:            "userID",
					SpotifyUserID: "spotify_user_id",
					DisplayName:   "display_name",
				}, nil)
			},
			want: &userRes{
				ID:          "userID",
				URI:         "spotify:user:spotify_user_id",
				DisplayName: "display_name",
				IsPremium:   false, // TODO : Spotifyの情報も正しく取ってこれるようにする
			},
			wantErr:  false,
			wantCode: http.StatusOK,
		},
		{
			name:   "ユーザの取得に失敗したときはInternalServerError",
			userID: "userID",
			prepareMockUserRepo: func(mock *mock_repository.MockUser) {
				mock.EXPECT().FindByID("userID").Return(nil, errors.New("unknown error"))
			},
			want:     nil,
			wantErr:  true,
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c = setToContext(c, tt.userID, nil)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mock := mock_repository.NewMockUser(ctrl)
			tt.prepareMockUserRepo(mock)
			uc := usecase.NewUserUseCase(nil, mock)
			h := &UserHandler{userUC: uc}

			err := h.GetMe(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMe() error = %v, wantErr %v", err, tt.wantErr)
			}
			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("GetMe() code = %d, want = %d", rec.Code, tt.wantCode)
			}

			if !tt.wantErr {
				got := &userRes{}
				err := json.Unmarshal(rec.Body.Bytes(), got)
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, tt.want) {
					t.Errorf("GetMe() diff = %v", cmp.Diff(got, tt.want))
				}
			}
		})
	}
}

func TestUserHandler_GetActiveDevices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		prepareMockUserSpo func(mock *mock_spotify.MockUser)
		want               *devicesRes
		wantErr            bool
		wantCode           int
	}{
		{
			name: "正しくデバイスを取得できる",
			prepareMockUserSpo: func(mock *mock_spotify.MockUser) {
				devices := []*entity.Device{
					{
						ID:           "hoge_id",
						IsRestricted: false,
						Name:         "hogeさんのiPhone11",
					},
				}
				mock.EXPECT().GetActiveDevices(gomock.Any()).Return(devices, nil)
			},
			want: &devicesRes{
				Devices: []*deviceJSON{
					{
						ID:           "hoge_id",
						IsRestricted: false,
						Name:         "hogeさんのiPhone11",
					},
				},
			},
			wantErr:  false,
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mock := mock_spotify.NewMockUser(ctrl)
			tt.prepareMockUserSpo(mock)
			uc := usecase.NewUserUseCase(mock, nil)
			h := &UserHandler{userUC: uc}

			err := h.GetActiveDevices(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetActiveDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("GetActiveDevices() code = %d, want = %d", rec.Code, tt.wantCode)
			}

			if !tt.wantErr {
				got := &devicesRes{}
				err := json.Unmarshal(rec.Body.Bytes(), got)
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, tt.want) {
					t.Errorf("GetActiveDevices() diff = %v", cmp.Diff(got, tt.want))
				}
			}
		})
	}
}
