package database

import (
	"fmt"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/repository"

	"github.com/go-gorp/gorp/v3"
	"github.com/go-sql-driver/mysql"
)

var _ repository.User = &UserRepository{}

// UserRepository は repository.UserRepository を満たす構造体です
type UserRepository struct {
	dbMap *gorp.DbMap
}

// NewUserRepository はUserRepositoryのポインタを生成する関数です
func NewUserRepository(dbMap *gorp.DbMap) *UserRepository {
	dbMap.AddTableWithName(userDTO{}, "users")
	return &UserRepository{dbMap: dbMap}
}

// FindByID は指定されたIDを持つユーザをDBから取得します
func (r *UserRepository) FindByID(id string) (*entity.User, error) {
	return &entity.User{
		ID: id,
	}, nil // TODO : 実際にDBから取得する
}

// StoreORUpdate はユーザを新規保存します。
func (r *UserRepository) Store(user *entity.User) error {
	dto := &userDTO{
		ID:            user.ID,
		SpotifyUserID: user.SpotifyUserID,
		DisplayName:   user.DisplayName,
	}
	if err := r.dbMap.Insert(dto); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return fmt.Errorf("insert user: %w", entity.ErrUserAlreadyExisted)
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

type userDTO struct {
	ID            string `db:"id"`
	SpotifyUserID string `db:"spotify_user_id"`
	DisplayName   string `db:"display_name"`
}
