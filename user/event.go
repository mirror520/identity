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

func ParseEventName(s string) EventName {
	switch s {
	case "user_registered":
		return UserRegistered
	case "user_activated":
		return UserActivated
	case "user_social_account_added":
		return UserSocialAccountAdded
	default:
		return Unknown
	}
}

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

func (name *EventName) MarshalJSON() ([]byte, error) {
	jsonStr := `"` + name.String() + `"`
	return []byte(jsonStr), nil
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
	name := strings.TrimPrefix(e.Name.String(), "user_")
	return "users." + e.UserID.String() + "." + name
}

type UserRegisteredEvent struct {
	*Event
	User *User `json:"user"`
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
	Account *SocialAccount `json:"account"`
}

func NewUserSocialAccountAddedEvent(u *User, account *SocialAccount) events.DomainEvent {
	return &UserSocialAccountAddedEvent{
		Event:   NewEvent(UserSocialAccountAdded, u),
		Account: account,
	}
}
