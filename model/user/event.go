package user

import (
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
	Domain string    `json:"domain"`
	Name   EventName `json:"name"`
	UserID UserID    `json:"user_id"`
	Time   time.Time `json:"time"`
}

func NewEvent(name EventName, id UserID) *Event {
	return &Event{
		Domain: "identity:users",
		Name:   name,
		UserID: id,
		Time:   time.Now(),
	}
}

func (e *Event) EventName() string {
	return e.Name.String()
}

type UserRegisteredEvent struct {
	*Event
	*User
}

func NewUserRegisteredEvent(u *User) events.DomainEvent {
	return &UserRegisteredEvent{
		Event: NewEvent(UserRegistered, u.ID),
		User:  u,
	}
}

type UserActivatedEvent struct {
	*Event
	Status Status `json:"status"`
}

func NewUserActivatedEvent(id UserID, status Status) events.DomainEvent {
	return &UserActivatedEvent{
		Event:  NewEvent(UserActivated, id),
		Status: status,
	}
}

type UserSocialAccountAddedEvent struct {
	*Event
	*SocialAccount
}

func NewUserSocialAccountAddedEvent(id UserID, account *SocialAccount) events.DomainEvent {
	return &UserSocialAccountAddedEvent{
		Event:         NewEvent(UserSocialAccountAdded, id),
		SocialAccount: account,
	}
}
