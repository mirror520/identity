package identity

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/api/idtoken"

	"github.com/mirror520/identity/model"
	"github.com/mirror520/identity/model/user"
)

var (
	ErrProviderNotSupported = errors.New("provider not supported")
	ErrClientIDNotFound     = errors.New("client id not found")
	ErrEmailNotFound        = errors.New("email not found")
	ErrNameNotFound         = errors.New("name not found")
	ErrPictureNotFound      = errors.New("picture not found")
)

type Service interface {
	SignIn(credential string, provider user.SocialProvider) (*user.User, error)
}

type ServiceMiddleware func(Service) Service

type service struct {
	users     user.Repository
	clientIDs map[user.SocialProvider]string
}

func NewService(users user.Repository, cfg model.Providers) Service {
	svc := new(service)
	svc.users = users
	svc.clientIDs = map[user.SocialProvider]string{
		user.GOOGLE: cfg.Google.Client.ID,
	}
	return svc
}

func (svc *service) SignIn(credential string, provider user.SocialProvider) (*user.User, error) {
	switch provider {
	case user.GOOGLE:
		return svc.signInWithGoogle(credential)
	}

	return nil, ErrProviderNotSupported
}

func (svc *service) signInWithGoogle(token string) (*user.User, error) {
	clientID, ok := svc.clientIDs[user.GOOGLE]
	if !ok {
		return nil, ErrClientIDNotFound
	}

	payload, err := idtoken.Validate(context.Background(), token, clientID)
	if err != nil {
		return nil, err
	}

	socialID := user.SocialAccountID(payload.Subject)
	u, err := svc.users.FindBySocialID(socialID)
	if err != nil {
		if !errors.Is(err, user.ErrUserNotFound) {
			return nil, err
		}

		// New User
		email, ok := payload.Claims["email"].(string)
		if !ok {
			return nil, ErrEmailNotFound
		}

		name, ok := payload.Claims["name"].(string)
		if !ok {
			return nil, ErrNameNotFound
		}

		username := strings.Split(email, "@")[0]

		u = user.NewUser(username, name, email)
		u.AddSocialAccount(user.GOOGLE, socialID)

		err := svc.users.Store(u) // Create User
		if err != nil {
			return nil, err
		}

		// Default Workspace
		w := u.DefaultWorkspace()
		err = svc.users.StoreWorkspace(w)
		if err != nil {
			return nil, err
		}
	}

	picture, ok := payload.Claims["picture"].(string)
	if !ok {
		return nil, ErrPictureNotFound
	}

	u.Avatar = picture
	return u, nil
}
