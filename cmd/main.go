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
	"github.com/mirror520/identity/pubsub/nats"
	"github.com/mirror520/identity/transport/http"
	"github.com/mirror520/identity/transport/pubsub"
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
		log.Fatal(err.Error())
	}
	defer log.Sync()

	zap.ReplaceGlobals(log)

	pubSub, err := nats.NewPullBasedPubSub(cfg.EventBus)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer pubSub.Close()

	stream := cfg.EventBus.Users.Stream
	if err := pubSub.AddStream(stream.Name, stream.Config); err != nil {
		log.Fatal(err.Error())
	}

	consumer := cfg.EventBus.Users.Consumer
	if err := pubSub.AddConsumer(consumer.Name, stream.Name, consumer.Config); err != nil {
		log.Fatal(err.Error())
	}

	repo, err := persistent.NewUserRepository(cfg.Persistent)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer repo.Close()

	r := gin.Default()
	r.Use(cors.Default())

	svc := identity.NewService(repo, cfg.Providers)
	svc = identity.LoggingMiddleware(log)(svc)

	// PATCH /signin
	{
		endpoint := identity.SignInEndpoint(svc)
		authenticator := http.SignInAuthenticator(endpoint)
		authMiddleware, err := http.AuthMiddlware(authenticator, *cfg)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		r.PATCH("/signin", authMiddleware.LoginHandler)
	}

	// POST /users
	{
		endpoint := identity.RegisterEndpoint(svc)
		r.POST("/users", http.RegisterHandler(endpoint))
	}

	// PATCH /users/:id/verify
	{
		endpoint := identity.OTPVerifyEndpoint(svc)
		r.POST("/users/:id/verify", http.OTPVerifyHandler(endpoint))
	}

	// PUT /users/id/socials
	{
		endpoint := identity.AddSocialAccountEndpoint(svc)
		r.POST("/users/:id/socials", http.AddSocialAccountHandler(endpoint))
	}

	// SUB users.>
	{
		endpoint := identity.EventEndpoint(svc)
		pubSub.PullSubscribe(consumer.Name, stream.Name, pubsub.EventHandler(endpoint))
	}

	r.Run(":8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sign := <-quit
	log.Info(sign.String())
}
