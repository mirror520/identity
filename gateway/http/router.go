package http

import (
	"github.com/gin-gonic/gin"

	jwt "github.com/appleboy/gin-jwt/v2"
)

func SetRouter(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) {
	apiV1 := r.Group("/v1")
	{
		apiV1.PATCH("/login", authMiddleware.LoginHandler)
	}
}
