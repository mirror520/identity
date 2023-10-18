package identity

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	"github.com/mirror520/identity/user"
)

type EndpointSet struct {
	Register         endpoint.Endpoint
	SignIn           endpoint.Endpoint
	OTPVerify        endpoint.Endpoint
	AddSocialAccount endpoint.Endpoint
	CheckHealth      endpoint.Endpoint
}

type RegisterRequest struct {
	Username string
	Name     string
	Email    string
}

func RegisterEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (response any, err error) {
		req, ok := request.(RegisterRequest)
		if !ok {
			return nil, errors.New("invalid request")
		}

		u, err := svc.Register(req.Username, req.Name, req.Email)
		if err != nil {
			return nil, err
		}

		return u, nil
	}
}

type OTPVerifyRequest struct {
	OTP    string
	UserID user.UserID
}

func OTPVerifyEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (response any, err error) {
		req, ok := request.(OTPVerifyRequest)
		if !ok {
			return nil, errors.New("invalid request")
		}

		u, err := svc.OTPVerify(req.OTP, req.UserID)
		if err != nil {
			return nil, err
		}

		return u, nil
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

		u, err := svc.SignIn(req.Credential, req.Provider)
		if err != nil {
			return nil, err
		}

		return u, nil
	}
}

type AddSocialAccountRequest struct {
	Credential string
	Provider   user.SocialProvider
	UserID     user.UserID
}

func AddSocialAccountEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (response any, err error) {
		req, ok := request.(AddSocialAccountRequest)
		if !ok {
			return nil, errors.New("invalid request")
		}

		u, err := svc.AddSocialAccount(req.Credential, req.Provider, req.UserID)
		if err != nil {
			return nil, err
		}

		return u, nil
	}
}

type RequestInfo struct {
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`
}

func CheckHealth(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (response any, err error) {
		err = svc.CheckHealth(ctx)
		return
	}
}

func EventEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (response any, err error) {
		handler, err := svc.Handler()
		if err != nil {
			return nil, err
		}

		switch e := request.(type) {
		case *user.UserRegisteredEvent:
			err = handler.UserRegisteredHandler(e)
		case *user.UserActivatedEvent:
			err = handler.UserActivatedHandler(e)
		case *user.UserSocialAccountAddedEvent:
			err = handler.UserSocialAccountAddedHandler(e)
		default:
			err = errors.New("invalid request")
		}

		return nil, err
	}
}
