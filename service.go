package identity

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/api/idtoken"

	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/user"
)

var (
	ErrProviderNotSupported = errors.New("provider not supported")
	ErrClientIDNotFound     = errors.New("client id not found")
	ErrEmailNotFound        = errors.New("email not found")
	ErrNameNotFound         = errors.New("name not found")
	ErrPictureNotFound      = errors.New("picture not found")
)

type Service interface {
	Register(username string, name string, email string) (*user.User, error)
	OTPVerify(otp string, id user.UserID) (*user.User, error)
	SignIn(credential string, provider user.SocialProvider) (*user.User, error)
	AddSocialAccount(credential string, provider user.SocialProvider, id user.UserID) (*user.User, error)

	UserRegisteredHandler(e *user.UserRegisteredEvent) error
	UserActivatedHandler(e *user.UserActivatedEvent) error
	UserSocialAccountAddedHandler(e *user.UserSocialAccountAddedEvent) error
}

type ServiceMiddleware func(Service) Service

type service struct {
	users     user.Repository
	clientIDs map[user.SocialProvider]string
}

func NewService(users user.Repository, cfg conf.Providers) Service {
	svc := new(service)
	svc.users = users
	svc.clientIDs = map[user.SocialProvider]string{
		user.GOOGLE: cfg.Google.Client.ID,
	}
	return svc
}

func (svc *service) Register(username string, name string, email string) (*user.User, error) {
	_, err := svc.users.FindByUsername(username)
	if err == nil {
		return nil, errors.New("user exists")
	}

	u := user.NewUser(username, name, email)
	defer u.Notify()

	return u, nil
}

func (svc *service) OTPVerify(otp string, id user.UserID) (*user.User, error) {
	u, err := svc.users.Find(id)
	if err != nil {
		return nil, err
	}

	// TODO: otp verify
	u.Activate()
	defer u.Notify()

	return u, nil
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

	socialID := user.SocialID(payload.Subject)
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

		defer u.Notify()
	}

	picture, ok := payload.Claims["picture"].(string)
	if ok {
		u.Avatar = picture
	}

	return u, nil
}

func (svc *service) AddSocialAccount(credential string, provider user.SocialProvider, id user.UserID) (*user.User, error) {
	u, err := svc.users.Find(id)
	if err != nil {
		return nil, err
	}

	clientID, ok := svc.clientIDs[user.GOOGLE]
	if !ok {
		return nil, ErrClientIDNotFound
	}

	payload, err := idtoken.Validate(context.Background(), credential, clientID)
	if err != nil {
		return nil, err
	}

	socialID := user.SocialID(payload.Subject)
	_, err = svc.users.FindBySocialID(socialID)
	if err == nil {
		return nil, errors.New("account exists")
	}

	u.AddSocialAccount(provider, socialID)
	defer u.Notify()

	return u, nil
}

func (svc *service) UserRegisteredHandler(e *user.UserRegisteredEvent) error {
	return svc.users.Store(e.User)
}

func (svc *service) UserActivatedHandler(e *user.UserActivatedEvent) error {
	u, err := svc.users.Find(e.UserID)
	if err != nil {
		return err
	}

	u.Status = e.Status
	u.UpdatedAt = e.OccuredAt

	return svc.users.Store(u)
}

func (svc *service) UserSocialAccountAddedHandler(e *user.UserSocialAccountAddedEvent) error {
	u, err := svc.users.Find(e.UserID)
	if err != nil {
		return err
	}

	u.Accounts = append(u.Accounts, e.Account)
	u.UpdatedAt = e.OccuredAt

	return svc.users.Store(u)
}
