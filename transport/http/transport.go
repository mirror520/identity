package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/endpoint"

	"github.com/mirror520/identity"
	"github.com/mirror520/identity/model"
	"github.com/mirror520/identity/model/user"
)

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

		result := model.SuccessResult("login success")
		result.Data = resp
		ctx.JSON(http.StatusOK, result)
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
