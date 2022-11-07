package identity

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mirror520/jinte/gateway"
	"github.com/mirror520/jinte/model"
	"github.com/mirror520/jinte/model/user"

	middleware "github.com/mirror520/jinte/gateway/http"
)

func SignInHandler(endpoint gateway.Endpoint) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req SignInRequest
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

func Authenticator(endpoint gateway.Endpoint) middleware.Authenticator {
	return func(ctx *gin.Context) (any, error) {
		var req SignInRequest
		if err := ctx.ShouldBind(&req); err != nil {
			return nil, err
		}

		resp, err := endpoint(ctx, req)
		if err != nil {
			return nil, err
		}

		user, ok := resp.(*user.User)
		if !ok {
			return nil, middleware.ErrFailedAuthentication
		}

		ctx.Set("user", user)

		return user, nil
	}
}
