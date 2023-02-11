package user

import (
	"errors"
	"time"

	"github.com/mirror520/identity/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserID uint

type User struct {
	ID       UserID             `json:"id" gorm:"primarykey"`
	Username string             `json:"username"`
	Name     string             `json:"name"`
	Email    string             `json:"email"`
	Accounts []*SocialAccount   `json:"accounts"`
	Members  []*WorkspaceMember `json:"members"`
	model.Time

	Avatar string `json:"avatar" gorm:"-"`
	Token  Token  `json:"token" gorm:"-"`
}

func NewUser(username string, name string, email string) *User {
	return &User{
		Username: username,
		Name:     name,
		Email:    email,
	}
}

func (u *User) AddSocialAccount(provider SocialProvider, id SocialAccountID) {
	if u.Accounts == nil {
		u.Accounts = make([]*SocialAccount, 0)
	}
	u.Accounts = append(u.Accounts, NewSocialAccount(provider, id))
}

func (u *User) DefaultWorkspace() *Workspace {
	return u.BuildWorkspace(u.Name + "'s Workspace")
}

func (u *User) BuildWorkspace(name string) *Workspace {
	w := NewWorkspace(name)
	w.AddMember(u, WorkspaceOwner)
	return w
}

type SocialProvider string

const (
	GOOGLE   SocialProvider = "google"
	FACEBOOK SocialProvider = "facebook"
	LINE     SocialProvider = "line"
)

type SocialAccountID string

type SocialAccount struct {
	UserID   UserID          `json:"user_id" gorm:"primarykey"`
	SocialID SocialAccountID `json:"id" gorm:"primarykey"`
	Provider SocialProvider  `json:"provider"`
	model.Time
}

func NewSocialAccount(provider SocialProvider, id SocialAccountID) *SocialAccount {
	return &SocialAccount{
		SocialID: id,
		Provider: provider,
	}
}

type Token struct {
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}
