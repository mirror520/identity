package identity

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	"github.com/mirror520/identity/model/user"
)

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
