package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	jwt "github.com/appleboy/gin-jwt/v2"

	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/model"
	"github.com/mirror520/identity/user"
)

var (
	ErrFailedAuthentication = jwt.ErrFailedAuthentication
)

type Authenticator func(*gin.Context) (any, error)

func AuthMiddlware(authenticator Authenticator, cfg conf.Config) (*jwt.GinJWTMiddleware, error) {
	identityKey := "username"

	mw := &jwt.GinJWTMiddleware{
		Realm:       cfg.BaseURL,
		Key:         []byte(cfg.JWT.Secret),
		Timeout:     cfg.JWT.Timeout,
		MaxRefresh:  cfg.JWT.Refresh.Maximum,
		IdentityKey: identityKey,

		// 身分驗證處理
		Authenticator: authenticator,

		// 附加 JWT Payloads
		PayloadFunc: func(data any) jwt.MapClaims {
			if v, ok := data.(*user.User); ok {
				return jwt.MapClaims{
					identityKey: v.Username,
				}
			}
			return jwt.MapClaims{}
		},

		// 登入成功之回應處理
		// from mw.LoginHandler
		LoginResponse: func(ctx *gin.Context, code int, token string, time time.Time) {
			v, _ := ctx.Get("user")
			u, ok := v.(*user.User)
			if !ok {
				err := errors.New("user not existed or type assert failure")
				result := model.FailureResult(err)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, result)
			}

			u.Token = user.Token{
				Token:     token,
				ExpiredAt: time,
			}

			result := model.SuccessResult("signin success")
			result.Data = u
			ctx.JSON(http.StatusOK, result)
		},

		// 授權驗證處理
		Authorizator: func(data interface{}, ctx *gin.Context) bool {
			// TODO: using OPA (Open Policy Agent)
			return true
		},

		// 更新 JWT Token 成功之回應處理
		// from mw.RefreshHandler
		RefreshResponse: func(ctx *gin.Context, code int, token string, time time.Time) {
			newToken := user.Token{
				Token:     token,
				ExpiredAt: time,
			}

			result := model.SuccessResult("refresh token success")
			result.Data = newToken
			ctx.JSON(http.StatusOK, result)
		},

		// 任何未取得授權之錯誤處理
		// from mw.unauthorized
		Unauthorized: func(ctx *gin.Context, code int, message string) {
			err := errors.New(message)
			result := model.FailureResult(err)
			ctx.JSON(code, result)
		},
	}

	return jwt.New(mw)
}
