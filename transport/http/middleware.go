package http

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/policy"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

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

type Claims struct {
	jwt.RegisteredClaims
	Roles []string `json:"roles"`
}

type GinAuth func(rule string) gin.HandlerFunc

func Authorizator(policy policy.Policy) GinAuth {
	return func(rule string) gin.HandlerFunc {
		cfg := conf.G()
		rules := strings.Split(rule, ".")
		domain := rules[0]
		action := rules[1]

		return func(ctx *gin.Context) {
			tokenStr := ctx.GetHeader("Authorization")

			if !strings.HasPrefix(tokenStr, "Bearer ") {
				unauthorized(ctx, http.StatusUnauthorized, ErrInvalidToken)
				return
			}
			tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

			var claims Claims
			_, err := jwt.ParseWithClaims(tokenStr, &claims, KeyFn(),
				jwt.WithIssuer(cfg.BaseURL),
				jwt.WithExpirationRequired(),
			)
			if err != nil {
				unauthorized(ctx, http.StatusUnauthorized, err)
				return
			}

			input := map[string]any{
				"domain": domain,
				"action": action,
				"roles":  claims.Roles,
			}

			allowed, err := policy.Eval(ctx, input)
			if err != nil {
				unauthorized(ctx, http.StatusExpectationFailed, err)
				return
			}

			if !allowed {
				unauthorized(ctx, http.StatusForbidden, errors.New("forbidden"))
				return
			}

			ctx.Next()
		}
	}
}
