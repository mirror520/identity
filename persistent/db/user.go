package db

import (
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/mirror520/identity/model/conf"
	"github.com/mirror520/identity/model/user"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(cfg conf.DB) (user.Repository, error) {
	db, err := gorm.Open(sqlite.Open(cfg.Name+".sql"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(
		&user.User{}, &user.SocialAccount{},
	)

	repo := new(userRepository)
	repo.db = db
	return repo, nil
}

func (repo *userRepository) Store(u *user.User) error {
	result := repo.db.Save(u)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (repo *userRepository) Find(id user.UserID) (*user.User, error) {
	var u *user.User

	result := repo.db.
		Preload("Accounts").
		Joins("INNER JOIN social_accounts ON social_accounts.user_id = users.id").
		Take(&u, "users.id = ? AND social_accounts.deleted_at IS NULL", id.String())

	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrUserNotFound
		}

		return nil, err
	}

	return u, nil
}

func (repo *userRepository) FindByUsername(username string) (*user.User, error) {
	var u *user.User

	result := repo.db.
		Preload("Accounts").
		Joins("INNER JOIN social_accounts ON social_accounts.user_id = users.id").
		Take(&u, "users.username = ? AND social_accounts.deleted_at IS NULL", username)

	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrUserNotFound
		}

		return nil, err
	}

	return u, nil
}

func (repo *userRepository) FindBySocialID(socialID string) (*user.User, error) {
	var u *user.User
	result := repo.db.
		Preload("Accounts").
		Joins("INNER JOIN social_accounts ON social_accounts.user_id = users.id").
		Take(&u, "social_accounts.social_id = ? AND social_accounts.deleted_at IS NULL", socialID)

	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrUserNotFound
		}

		return nil, err
	}

	return u, nil
}

func (repo *userRepository) DB() *gorm.DB {
	return repo.db
}
