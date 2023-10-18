package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-kit/kit/endpoint"

	"github.com/mirror520/identity"
	"github.com/mirror520/identity/model"
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

func SignInHandler(endpoint endpoint.Endpoint) pubsub.MessageHandler {
	return func(ctx context.Context, msg *pubsub.Message) error {
		var req identity.SignInRequest

		if err := json.Unmarshal(msg.Data, &req); err != nil {
			result := model.FailureResult(err)
			bs, err := result.Bytes()
			if err != nil {
				return err
			}
			return msg.Response(bs)
		}

		resp, err := endpoint(ctx, req)
		if err != nil {
			result := model.FailureResult(err)
			bs, err := result.Bytes()
			if err != nil {
				return err
			}
			return msg.Response(bs)
		}

		result := model.SuccessResult("user signed in")
		result.Data = resp

		bs, err := result.Bytes()
		if err != nil {
			return err
		}

		return msg.Response(bs)
	}
}

func CheckHealthHandler(endpoint endpoint.Endpoint) pubsub.MessageHandler {
	return func(_ context.Context, msg *pubsub.Message) error {
		var info *identity.RequestInfo
		if err := json.Unmarshal(msg.Data, &info); err != nil {
			return err
		}

		ctx := context.WithValue(context.Background(), model.REQUEST_INFO, info)
		_, err := endpoint(ctx, nil)
		if err != nil {
			return msg.Response([]byte(err.Error()))
		}

		return msg.Response([]byte(`ok`))
	}
}
