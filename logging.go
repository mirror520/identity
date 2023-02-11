package identity

import (
	"go.uber.org/zap"

	"github.com/mirror520/identity/model/user"
)

func LoggingMiddleware() ServiceMiddleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			zap.L().With(
				zap.String("service", "identity"),
				zap.String("middleware", "logging"),
			),
			next,
		}
	}
}

type loggingMiddleware struct {
	log  *zap.Logger
	next Service
}

func (mw *loggingMiddleware) SignIn(credential string, provider user.SocialProvider) (*user.User, error) {
	log := mw.log.With(
		zap.String("action", "signin"),
	)

	u, err := mw.next.SignIn(credential, provider)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	log.Info("user signin",
		zap.String("username", u.Username),
	)
	return u, nil
}
