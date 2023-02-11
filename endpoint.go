package identity

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"github.com/mirror520/identity/model/user"
)

type SignInRequest struct {
	Credential string
	Provider   user.SocialProvider
}

func SignInEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (response any, err error) {
		req := request.(SignInRequest)
		user, err := svc.SignIn(req.Credential, req.Provider)
		if err != nil {
			return nil, err
		}

		return user, nil
	}
}
