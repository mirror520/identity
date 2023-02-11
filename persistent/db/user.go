package db

import (
	"errors"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/mirror520/identity/model"
	"github.com/mirror520/identity/model/user"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository() (user.Repository, error) {
	cfg := model.Config
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(
		&user.Workspace{}, &user.WorkspaceMember{},
		&user.User{}, &user.SocialAccount{},
	)

	repo := new(userRepository)
	repo.db = db
	return repo, nil
}

func (repo *userRepository) Store(u *user.User) error {
	var result *gorm.DB

	if u.ID == 0 {
		result = repo.db.Create(u)
	} else {
		result = repo.db.Save(u)
	}

	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (repo *userRepository) FindBySocialID(id user.SocialAccountID) (*user.User, error) {
	var u *user.User
	result := repo.db.
		Preload("Accounts").
		Joins("INNER JOIN social_accounts ON social_accounts.user_id = users.id").
		Take(&u, "social_accounts.social_id = ? AND social_accounts.deleted_at IS NULL", id)

	err := result.Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrUserNotFound
		}

		return nil, err
	}

	return u, nil
}

func (repo *userRepository) StoreWorkspace(w *user.Workspace) error {
	var result *gorm.DB

	if w.ID == 0 {
		result = repo.db.Create(w)
	} else {
		result = repo.db.Save(w)
	}

	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (repo *userRepository) FindWorkspaces(id user.UserID) ([]*user.WorkspaceMember, error) {
	var members []*user.WorkspaceMember // member includes workspace and user
	result := repo.db.
		Preload("Workspace").
		Find(&members, "workspace_members.user_id = ? AND workspace_members.deleted_at IS NULL", id)

	err := result.Error
	if err != nil {
		return nil, err
	}

	return members, nil
}

func (repo *userRepository) DB() *gorm.DB {
	return repo.db
}
