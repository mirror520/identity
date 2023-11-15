package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mirror520/identity/policy"
)

type Claims struct {
	jwt.RegisteredClaims
	Roles []string `json:"roles"`
}

func (c *Claims) Map() map[string]any {
	return map[string]any{
		"sub":   c.Subject,
		"roles": c.Roles,
	}
}

type Who byte

const (
	Owner Who = 1 << iota
	Group
	Others
	Admin
	All
)

type GinAuth func(rule string, who ...Who) gin.HandlerFunc

func Authorizator(policy policy.Policy) GinAuth {
	return func(rule string, who ...Who) gin.HandlerFunc {
		rules := strings.Split(rule, ".")
		domain := rules[0]
		action := rules[1]

		var flags byte
		for _, w := range who {
			flags = flags | byte(w)
		}

		return func(ctx *gin.Context) {
			var claims Claims
			if err := ParseToken(ctx, &claims); err != nil {
				unauthorized(ctx, http.StatusUnauthorized, err)
				return
			}

			input := map[string]any{
				"domain":    domain,
				"action":    action,
				"who_flags": flags,
				"claims":    claims.Map(),
			}

			if id := ctx.Param("id"); id != "" {
				input["object"] = id
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
