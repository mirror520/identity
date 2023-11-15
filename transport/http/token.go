package http

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mirror520/identity/conf"
)

var ErrInvalidToken = errors.New("invalid token")

var (
	keyFn jwt.Keyfunc
	once  sync.Once
)

func KeyFn() jwt.Keyfunc {
	once.Do(func() {
		secret := conf.G().JWT.Secret
		keyFn = func(t *jwt.Token) (interface{}, error) {
			return secret, nil
		}
	})

	return keyFn
}

func ParseToken(ctx *gin.Context, claims jwt.Claims) error {
	tokenStr := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(tokenStr, "Bearer ") {
		return ErrInvalidToken
	}

	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	_, err := jwt.ParseWithClaims(tokenStr, claims, KeyFn(),
		jwt.WithIssuer(conf.G().BaseURL),
		jwt.WithLeeway(10*time.Second),
	)

	return err
}
