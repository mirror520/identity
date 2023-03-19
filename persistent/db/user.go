package db

import (
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/user"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(cfg conf.Persistent) (user.Repository, error) {
	filename := cfg.Name + ".db"
	if cfg.InMem {
		filename = "file::memory:?cache=shared"
	}

	db, err := gorm.Open(sqlite.Open(filename), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(
		&User{}, &SocialAccount{},
	)

	repo := new(userRepository)
	repo.db = db
	return repo, nil
}

func (repo *userRepository) Store(u *user.User) error {
	user := NewUser(u) // convert Domain to Data model

	result := repo.db.Save(user)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (repo *userRepository) Find(id user.UserID) (*user.User, error) {
	var u *User

	result := repo.db.
		Preload("Accounts").
		Joins("LEFT JOIN social_accounts ON social_accounts.user_id = users.id").
		Take(&u, "users.id = ? AND social_accounts.deleted_at IS NULL", id.String())

	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrUserNotFound
		}

		return nil, err
	}

	user := u.reconstitute()
	return user, nil
}

func (repo *userRepository) FindByUsername(username string) (*user.User, error) {
	var u *User

	result := repo.db.
		Preload("Accounts").
		Joins("LEFT JOIN social_accounts ON social_accounts.user_id = users.id").
		Take(&u, "users.username = ? AND social_accounts.deleted_at IS NULL", username)

	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrUserNotFound
		}

		return nil, err
	}

	user := u.reconstitute()
	return user, nil
}

func (repo *userRepository) FindBySocialID(socialID user.SocialID) (*user.User, error) {
	var u *User
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

	user := u.reconstitute()
	return user, nil
}

func (repo *userRepository) Close() error {
	return nil
}

func (repo *userRepository) DB() *gorm.DB {
	return repo.db
}
