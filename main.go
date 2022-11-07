package main

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/configor"
	"go.uber.org/zap"

	"github.com/mirror520/jinte/gateway/http"
	"github.com/mirror520/jinte/model"
	"github.com/mirror520/jinte/persistent/db"
	"github.com/mirror520/jinte/service/identity"
)

func main() {
	os.Setenv("CONFIGOR_ENV_PREFIX", "JINTE")
	configor.Load(&model.Config, "config.yaml")

	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err.Error())
	}
	defer log.Sync()

	zap.ReplaceGlobals(log)

	users, err := db.NewUserRepository()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	var authenticator http.Authenticator
	{
		svc := identity.NewService(users)
		endpint := identity.SignInEndpoint(svc)
		authenticator = identity.Authenticator(endpint)
	}

	authMiddleware, err := http.AuthMiddlware(authenticator)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	r := gin.Default()
	r.Use(cors.Default())

	http.SetRouter(r, authMiddleware)

	r.Run(":8080")
}
