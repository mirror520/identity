package user

import (
	"errors"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/mirror520/identity/events"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Status int

const (
	Pending Status = iota
	Registered
	Activated
	Locked
	Revoked
)

type UserID ulid.ULID // AggregateRoot

func NewID() UserID {
	return UserID(ulid.Make())
}

func (id UserID) Bytes() []byte {
	return id[:]
}

func (id UserID) String() string {
	return ulid.ULID(id).String()
}

type User struct {
	ID       UserID           `json:"id"`
	Username string           `json:"username"`
	Name     string           `json:"name"`
	Email    string           `json:"email"`
	Status   Status           `json:"status"`
	Accounts []*SocialAccount `json:"accounts"`
	Avatar   string           `json:"avatar"`
	Token    Token            `json:"token"`

	events.EventStore `json:"-"`
}

func NewUser(username string, name string, email string) *User {
	u := &User{
		ID:       NewID(),
		Username: username,
		Name:     name,
		Email:    email,
		Status:   Pending,

		EventStore: events.NewEventStore(),
	}
	u.Register()

	return u
}

func (u *User) Register() {
	u.Status = Registered

	e := NewUserRegisteredEvent(u)
	u.AddEvent(e)
}

func (u *User) Activate() {
	u.Status = Activated

	e := NewUserActivatedEvent(u.ID, Activated)
	u.AddEvent(e)
}

func (u *User) AddSocialAccount(provider SocialProvider, socialID string) {
	account := NewSocialAccount(provider, socialID)

	if u.Accounts == nil {
		u.Accounts = make([]*SocialAccount, 0)
	}
	u.Accounts = append(u.Accounts, account)

	e := NewUserSocialAccountAddedEvent(u.ID, account)
	u.AddEvent(e)
}

type SocialProvider string

const (
	GOOGLE   SocialProvider = "google"
	FACEBOOK SocialProvider = "facebook"
	LINE     SocialProvider = "line"
)

type SocialAccount struct {
	SocialID string         `json:"social_id"`
	Provider SocialProvider `json:"social_provider"`
}

func NewSocialAccount(provider SocialProvider, socialID string) *SocialAccount {
	return &SocialAccount{
		SocialID: socialID,
		Provider: provider,
	}
}

type Token struct {
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}
