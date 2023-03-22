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
		if len(ss) != 2 || ss[0] != "users" {
			return errors.New("invalid router")
		}

		var event any
		switch ss[1] {
		case "registered":
			var e *user.UserRegisteredEvent
			if err := json.Unmarshal(msg.Data, &e); err != nil {
				return err
			}
			event = e

		case "activated":
			var e *user.UserRegisteredEvent
			if err := json.Unmarshal(msg.Data, &e); err != nil {
				return err
			}
			event = e

		case "social_account_added":
			var e *user.UserSocialAccountAddedEvent
			if err := json.Unmarshal(msg.Data, &e); err != nil {
				return err
			}
			event = e
		}

		_, err := endpoint(ctx, event)
		return err
	}
}
