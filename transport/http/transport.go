package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/endpoint"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"

	"github.com/mirror520/identity"
	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/model"
	"github.com/mirror520/identity/user"
)

func RegisterHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req identity.RegisterRequest
		if err := ctx.ShouldBind(&req); err != nil {
			result := model.FailureResult(err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, result)
			return
		}

		resp, err := endpoint(ctx, req)
		if err != nil {
			result := model.FailureResult(err)
			ctx.AbortWithStatusJSON(http.StatusForbidden, result)
			return
		}

		result := model.SuccessResult("user registered")
		result.Data = resp
		ctx.JSON(http.StatusOK, result)
	}
}

func OTPVerifyHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		if id == "" {
			err := errors.New("id not found")
			result := model.FailureResult(err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, result)
		}

		userID, err := user.ParseID(id)
		if err != nil {
			result := model.FailureResult(err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, result)
		}

		var req identity.OTPVerifyRequest
		if err := ctx.ShouldBind(&req); err != nil {
			result := model.FailureResult(err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, result)
			return
		}
		req.UserID = userID

		resp, err := endpoint(ctx, req)
		if err != nil {
			result := model.FailureResult(err)
			ctx.AbortWithStatusJSON(http.StatusForbidden, result)
			return
		}

		result := model.SuccessResult("user verified")
		result.Data = resp
		ctx.JSON(http.StatusOK, result)
	}
}

func SignInHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req identity.SignInRequest
		err := ctx.ShouldBind(&req)
		if err != nil {
			result := model.FailureResult(err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, result)
			return
		}

		resp, err := endpoint(ctx, req)
		if err != nil {
			result := model.FailureResult(err)
			ctx.AbortWithStatusJSON(http.StatusExpectationFailed, result)
			return
		}

		u, ok := resp.(*user.User)
		if !ok {
			err := errors.New("invalid user")
			unauthorized(ctx, http.StatusExpectationFailed, err)
			return
		}

		cfg := conf.G()
		now := time.Now()
		claims := Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    cfg.BaseURL,
				Subject:   u.Username,
				ExpiresAt: jwt.NewNumericDate(now.Add(cfg.JWT.Timeout)),
				IssuedAt:  jwt.NewNumericDate(now),
				ID:        ulid.Make().String(),
			},
			Roles: []string{"admin"},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
		tokenStr, err := token.SignedString(cfg.JWT.Secret)
		if err != nil {
			unauthorized(ctx, http.StatusExpectationFailed, err)
			return
		}

		u.Token = user.Token{
			Token:     tokenStr,
			ExpiredAt: now.Add(cfg.JWT.Timeout),
		}

		result := model.SuccessResult("user signed in")
		result.Data = resp
		ctx.JSON(http.StatusOK, result)
	}
}

func unauthorized(ctx *gin.Context, code int, err error) {
	realm := conf.G().BaseURL

	ctx.Abort()
	ctx.Header("WWW-Authenticate", "Bearer realm="+realm)
	ctx.String(code, err.Error())
}

func RefreshHandler(ctx *gin.Context) {
	cfg := conf.G()
	if !cfg.JWT.Refresh.Enabled {
		ctx.Abort()
		ctx.String(http.StatusForbidden, "token refresh disabled")
		return
	}

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

	if time.Since(claims.IssuedAt.Time) > cfg.JWT.Refresh.Maximum {
		err := errors.New("token beyond refresh time")
		unauthorized(ctx, http.StatusUnauthorized, err)
		return
	}

	now := time.Now()
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(cfg.JWT.Timeout))
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ID = ulid.Make().String()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	tokenStr, err = token.SignedString(cfg.JWT.Secret)
	if err != nil {
		unauthorized(ctx, http.StatusExpectationFailed, err)
		return
	}

	t := user.Token{
		Token:     tokenStr,
		ExpiredAt: now.Add(cfg.JWT.Timeout),
	}

	result := model.SuccessResult("token refreshed")
	result.Data = t
	ctx.JSON(http.StatusOK, result)
}

func AddSocialAccountHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		if id == "" {
			err := errors.New("id not found")
			result := model.FailureResult(err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, result)
		}

		userID, err := user.ParseID(id)
		if err != nil {
			result := model.FailureResult(err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, result)
		}

		var req identity.AddSocialAccountRequest
		if err := ctx.ShouldBind(&req); err != nil {
			result := model.FailureResult(err)
			ctx.AbortWithStatusJSON(http.StatusBadRequest, result)
			return
		}
		req.UserID = userID

		resp, err := endpoint(ctx, req)
		if err != nil {
			result := model.FailureResult(err)
			ctx.AbortWithStatusJSON(http.StatusForbidden, result)
			return
		}

		result := model.SuccessResult("user social account added")
		result.Data = resp
		ctx.JSON(http.StatusOK, result)
	}
}

func CheckHealthHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		info := &identity.RequestInfo{
			ClientIP:  c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}

		ctx := context.WithValue(context.Background(), model.REQUEST_INFO, info)
		_, err := endpoint(ctx, nil)
		if err != nil {
			result := model.FailureResult(err)
			c.AbortWithStatusJSON(http.StatusExpectationFailed, result)
			return
		}

		c.String(http.StatusOK, "ok")
	}
}
