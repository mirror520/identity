package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/mirror520/identity"
	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/persistent"
	"github.com/mirror520/identity/transport/http"
)

func main() {
	path, ok := os.LookupEnv("IDENTITY_PATH")
	if !ok {
		path = "."
	}

	cfg, err := conf.LoadConfig(path)
	if err != nil {
		log.Fatal(err.Error())
	}

	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err.Error())
	}
	defer log.Sync()

	zap.ReplaceGlobals(log)

	repo, err := persistent.NewUserRepository(cfg.Persistent)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer repo.Close()

	var authenticator http.Authenticator
	{
		svc := identity.NewService(repo, cfg.Providers)
		svc = identity.LoggingMiddleware(log)(svc)
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

	r.PATCH("/login", authMiddleware.LoginHandler)

	r.Run(":8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sign := <-quit
	log.Info(sign.String())
}
