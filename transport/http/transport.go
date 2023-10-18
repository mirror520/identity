package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/endpoint"

	"github.com/mirror520/identity"
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
			ctx.AbortWithStatusJSON(http.StatusForbidden, result)
			return
		}

		result := model.SuccessResult("user signed in")
		result.Data = resp
		ctx.JSON(http.StatusOK, result)
	}
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

func SignInAuthenticator(endpoint endpoint.Endpoint) Authenticator {
	return func(ctx *gin.Context) (any, error) {
		var req identity.SignInRequest
		if err := ctx.ShouldBind(&req); err != nil {
			return nil, err
		}

		resp, err := endpoint(ctx, req)
		if err != nil {
			return nil, err
		}

		user, ok := resp.(*user.User)
		if !ok {
			return nil, ErrFailedAuthentication
		}

		ctx.Set("user", user)

		return user, nil
	}
}
