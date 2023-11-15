package user

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/mirror520/identity/events"
	"github.com/mirror520/identity/model"
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

func ParseStatus(status string) (Status, error) {
	status = strings.ToLower(status)
	switch status {
	case "pending":
		return Pending, nil
	case "registered":
		return Registered, nil
	case "activated":
		return Activated, nil
	case "locked":
		return Locked, nil
	case "revoked":
		return Revoked, nil
	default:
		return -1, errors.New("invalid status")
	}
}

func (s Status) String() string {
	switch s {
	case Pending:
		return "pending"
	case Registered:
		return "registered"
	case Activated:
		return "activated"
	case Locked:
		return "locked"
	case Revoked:
		return "revoked"
	default:
		return "unknown"
	}
}

func (s *Status) MarshalJSON() ([]byte, error) {
	jsonStr := `"` + s.String() + `"`
	return []byte(jsonStr), nil
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	status, err := ParseStatus(raw)
	if err != nil {
		return err
	}

	*s = status
	return nil
}

type UserID ulid.ULID // AggregateRoot

func MakeID() UserID {
	return UserID(ulid.Make())
}

func ParseID(id string) (UserID, error) {
	userID, err := ulid.Parse(id)
	if err != nil {
		return UserID{}, err
	}
	return UserID(userID), nil
}

func (id UserID) Bytes() []byte {
	return id[:]
}

func (id UserID) String() string {
	return ulid.ULID(id).String()
}

func (id UserID) Time() time.Time {
	ms := ulid.ULID(id).Time()
	return ulid.Time(ms)
}

func (id *UserID) MarshalJSON() ([]byte, error) {
	jsonStr := `"` + id.String() + `"`
	return []byte(jsonStr), nil
}

func (id *UserID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	userID, err := ParseID(s)
	if err != nil {
		return err
	}

	*id = userID
	return nil
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
	model.Model

	events.EventStore `json:"-"`
}

func NewUser(username string, name string, email string) *User {
	id := MakeID()

	u := &User{
		ID:       id,
		Username: username,
		Name:     name,
		Email:    email,
		Status:   Pending,
		Model: model.Model{
			CreatedAt: id.Time(),
		},

		EventStore: events.NewEventStore(),
	}

	u.Register()
	return u
}

func (u *User) Register() {
	u.Status = Registered
	u.UpdatedAt = time.Now()

	e := NewUserRegisteredEvent(u)
	u.AddEvent(e)
}

func (u *User) Activate() {
	u.Status = Activated
	u.UpdatedAt = time.Now()

	e := NewUserActivatedEvent(u, Activated)
	u.AddEvent(e)
}

func (u *User) AddSocialAccount(provider SocialProvider, socialID SocialID) {
	account := NewSocialAccount(provider, socialID)

	if u.Accounts == nil {
		u.Accounts = make([]*SocialAccount, 0)
	}
	u.Accounts = append(u.Accounts, account)

	u.UpdatedAt = account.UpdatedAt

	e := NewUserSocialAccountAddedEvent(u, account)
	u.AddEvent(e)
}

type SocialProvider string

const (
	GOOGLE   SocialProvider = "google"
	FACEBOOK SocialProvider = "facebook"
	LINE     SocialProvider = "line"
)

type SocialID string

type SocialAccount struct {
	SocialID SocialID       `json:"social_id"`
	Provider SocialProvider `json:"social_provider"`
	model.Model
}

func NewSocialAccount(provider SocialProvider, id SocialID) *SocialAccount {
	return &SocialAccount{
		SocialID: id,
		Provider: provider,
		Model: model.Model{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

type Token struct {
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expired_at"`
}
