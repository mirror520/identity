package identity

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	"github.com/mirror520/identity/user"
)

type RegisterRequest struct {
	Username string
	Name     string
	Email    string
}

func RegisterHandler(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (response any, err error) {
		req, ok := request.(RegisterRequest)
		if !ok {
			return nil, errors.New("invalid request")
		}

		user, err := svc.Register(req.Username, req.Name, req.Email)
		if err != nil {
			return nil, err
		}

		return user, nil
	}
}

type SignInRequest struct {
	Credential string
	Provider   user.SocialProvider
}

func SignInEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (response any, err error) {
		req, ok := request.(SignInRequest)
		if !ok {
			return nil, errors.New("invalid request")
		}

		user, err := svc.SignIn(req.Credential, req.Provider)
		if err != nil {
			return nil, err
		}

		return user, nil
	}
}

func EventEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (response any, err error) {
		switch e := request.(type) {
		case *user.UserRegisteredEvent:
			err = svc.UserRegisteredHandler(e)
		case *user.UserActivatedEvent:
			err = svc.UserActivatedHandler(e)
		case *user.UserSocialAccountAddedEvent:
			err = svc.UserSocialAccountAddedHandler(e)
		default:
			err = errors.New("invalid request")
		}

		return nil, err
	}
}
