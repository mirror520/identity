package user

import (
	"strings"
	"time"

	"github.com/mirror520/identity/events"
)

type EventName int

const (
	Unknown EventName = iota
	UserRegistered
	UserActivated
	UserSocialAccountAdded
)

func (name EventName) String() string {
	switch name {
	case UserRegistered:
		return "user_registered"
	case UserActivated:
		return "user_activated"
	case UserSocialAccountAdded:
		return "user_social_account_added"
	default:
		return ""
	}
}

type Event struct {
	Domain    string    `json:"domain"`
	Name      EventName `json:"name"`
	UserID    UserID    `json:"user_id"` // AggreagateRoot
	OccuredAt time.Time `json:"occured_at"`
}

func NewEvent(name EventName, u *User) *Event {
	return &Event{
		Domain:    "identity:users",
		Name:      name,
		UserID:    u.ID,
		OccuredAt: u.UpdatedAt,
	}
}

func (e *Event) EventName() string {
	return e.Name.String()
}

func (e *Event) Topic() string {
	return strings.TrimPrefix(e.Name.String(), "user_")
}

type UserRegisteredEvent struct {
	*Event
	*User
}

func NewUserRegisteredEvent(u *User) events.DomainEvent {
	return &UserRegisteredEvent{
		Event: NewEvent(UserRegistered, u),
		User:  u,
	}
}

type UserActivatedEvent struct {
	*Event
	Status Status `json:"status"`
}

func NewUserActivatedEvent(u *User, status Status) events.DomainEvent {
	return &UserActivatedEvent{
		Event:  NewEvent(UserActivated, u),
		Status: status,
	}
}

type UserSocialAccountAddedEvent struct {
	*Event
	*SocialAccount
}

func NewUserSocialAccountAddedEvent(u *User, account *SocialAccount) events.DomainEvent {
	return &UserSocialAccountAddedEvent{
		Event:         NewEvent(UserSocialAccountAdded, u),
		SocialAccount: account,
	}
}
