package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/mirror520/identity"
	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/persistent/db"
	"github.com/mirror520/identity/transport/http"
)

func main() {
	cfg, err := conf.LoadConfig(".")
	if err != nil {
		log.Fatal(err.Error())
	}

	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err.Error())
	}
	defer log.Sync()

	zap.ReplaceGlobals(log)

	repo, err := db.NewUserRepository(cfg.DB)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	var authenticator http.Authenticator
	{
		svc := identity.NewService(repo, cfg.Providers)
		endpoint := identity.SignInEndpoint(svc)
		authenticator = http.SignInAuthenticator(endpoint)
	}

	authMiddleware, err := http.AuthMiddlware(authenticator, *cfg)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	r := gin.Default()
	r.Use(cors.Default())

	apiV1 := r.Group("/v1")
	{
		apiV1.PATCH("/login", authMiddleware.LoginHandler)
	}

	r.Run(":8080")
}
