package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-kit/kit/endpoint"

	"github.com/mirror520/identity/pubsub"
	"github.com/mirror520/identity/user"
)

func EventHandler(endpoint endpoint.Endpoint) pubsub.MessageHandler {
	return func(ctx context.Context, msg *pubsub.Message) error {
		ss := strings.Split(msg.Topic, ".")
		if len(ss) != 3 || ss[0] != "users" {
			return errors.New("invalid event")
		}

		name := user.ParseEventName("user_" + ss[2])

		var event any
		switch name {
		case user.UserRegistered:
			var e *user.UserRegisteredEvent
			if err := json.Unmarshal(msg.Data, &e); err != nil {
				return err
			}
			event = e

		case user.UserActivated:
			var e *user.UserActivatedEvent
			if err := json.Unmarshal(msg.Data, &e); err != nil {
				return err
			}
			event = e

		case user.UserSocialAccountAdded:
			var e *user.UserSocialAccountAddedEvent
			if err := json.Unmarshal(msg.Data, &e); err != nil {
				return err
			}
			event = e

		default:
			return errors.New("invalid event")
		}

		_, err := endpoint(ctx, event)
		return err
	}
}
