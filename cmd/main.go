package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	consul "github.com/hashicorp/consul/api"

	"github.com/mirror520/identity"
	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/events"
	"github.com/mirror520/identity/persistent"
	"github.com/mirror520/identity/pubsub/nats"
	"github.com/mirror520/identity/transport/http"
	"github.com/mirror520/identity/transport/pubsub"
)

var (
	Version   string
	BuildTime string
	GitCommit string
)

func main() {
	cli.VersionPrinter = func(cli *cli.Context) {
		fmt.Println("Version: " + cli.App.Version)
		fmt.Println("BuildTime: " + BuildTime)
		fmt.Println("GitCommit: " + GitCommit)
	}

	app := &cli.App{
		Name:    "identity",
		Usage:   "Scalable and decentralized user identity management",
		Version: Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "path",
				Usage:   "Specifies the working directory for the identity microservice.",
				EnvVars: []string{"IDENTITY_PATH"},
			},
			&cli.IntFlag{
				Name:    "port",
				Usage:   "Specifies the HTTP service port for the identity microservice.",
				Value:   8080,
				EnvVars: []string{"IDENTITY_HTTP_PORT"},
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	time.Sleep(3000 * time.Millisecond)
}

func run(cli *cli.Context) error {
	path := cli.String("path")
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		path = homeDir + "/.identity"
	}

	cfg, err := conf.LoadConfig(path)
	if err != nil {
		return err
	}
	cfg.Port = cli.Int("port")

	log, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	defer log.Sync()

	zap.ReplaceGlobals(log)

	log = log.With(zap.String("action", "main"))

	pubSub, err := nats.NewPullBasedPubSub(cfg.EventBus)
	if err != nil {
		log.Error(err.Error(), zap.String("infra", "nats"))
		return err
	}
	defer pubSub.Close()

	events.ReplaceGlobals(pubSub)

	stream := cfg.EventBus.Users.Stream
	if err := pubSub.AddStream(stream.Name, stream.Config); err != nil {
		log.Error(err.Error(),
			zap.String("infra", "nats"),
			zap.String("phase", "add_stream"),
			zap.String("stream", stream.Name),
		)
		return err
	}

	consumer := cfg.EventBus.Users.Consumer
	if err := pubSub.AddConsumer(consumer.Name, stream.Name, consumer.Config); err != nil {
		log.Error(err.Error(),
			zap.String("infra", "nats"),
			zap.String("phase", "add_consumer"),
			zap.String("stream", stream.Name),
			zap.String("consumer", consumer.Name),
		)
		return err
	}

	repo, err := persistent.NewUserRepository(cfg.Persistent)
	if err != nil {
		log.Error(err.Error(),
			zap.String("infra", "persistent"),
			zap.String("driver", cfg.Persistent.Driver.String()),
		)
		return err
	}
	defer repo.Close()

	r := gin.Default()
	r.Use(cors.Default())

	svc := identity.NewService(repo, cfg.Providers)
	svc = identity.LoggingMiddleware(log)(svc)

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, "ok")
	})

	apiV1 := r.Group("/identity/v1")
	{
		// PATCH /signin
		{
			endpoint := identity.SignInEndpoint(svc)
			authenticator := http.SignInAuthenticator(endpoint)
			authMiddleware, err := http.AuthMiddlware(authenticator, *cfg)
			if err != nil {
				return err
			}

			apiV1.PATCH("/signin", authMiddleware.LoginHandler)

			if cfg.Transport.NATS.Enabled {
				pubSub.Subscribe("identity.signin", pubsub.SignInHandler(endpoint))                      // NATS LB
				pubSub.Subscribe(cfg.Transport.NATS.ReqPrefix+".signin", pubsub.SignInHandler(endpoint)) // NATS Direct
			}
		}

		// POST /users
		{
			endpoint := identity.RegisterEndpoint(svc)
			apiV1.POST("/users", http.RegisterHandler(endpoint))
		}

		// PATCH /users/:id/verify
		{
			endpoint := identity.OTPVerifyEndpoint(svc)
			apiV1.POST("/users/:id/verify", http.OTPVerifyHandler(endpoint))
		}

		// PUT /users/id/socials
		{
			endpoint := identity.AddSocialAccountEndpoint(svc)
			apiV1.POST("/users/:id/socials", http.AddSocialAccountHandler(endpoint))
		}
	}

	// SUB users.>
	{
		endpoint := identity.EventEndpoint(svc)
		pubSub.PullSubscribe(consumer.Name, stream.Name, pubsub.EventHandler(endpoint))
	}

	go r.Run(":" + strconv.Itoa(cfg.Port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go Registry(ctx, cfg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sign := <-quit
	log.Info(sign.String())

	return nil
}

func Registry(ctx context.Context, cfg *conf.Config) {
	log := zap.L().With(
		zap.String("action", "service_registry"),
	)

	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		log.Error(err.Error())
		return
	}

	http := "http"
	address := "localhost"
	port := cfg.Port

	if cfg.ExternalProxy != nil {
		http = cfg.ExternalProxy.Scheme
		address = cfg.ExternalProxy.Address
		port = cfg.ExternalProxy.Port
	}

	service := &consul.AgentServiceRegistration{
		ID:      cfg.Name,
		Name:    "identity",
		Tags:    []string{http},
		Port:    port,
		Address: address,
		TaggedAddresses: map[string]consul.ServiceAddress{
			http: {Address: address, Port: port},
		},
		Meta: make(map[string]string),
		Check: &consul.AgentServiceCheck{
			Interval:                       "10s",
			Timeout:                        "1s",
			HTTP:                           http + "://" + address + ":" + strconv.Itoa(port) + "/health",
			DeregisterCriticalServiceAfter: "60s",
		},
	}

	if cfg.Transport.NATS.Enabled {
		service.Tags = append(service.Tags, "nats")
		service.TaggedAddresses["nats"] = consul.ServiceAddress{
			Address: cfg.EventBus.Host,
			Port:    cfg.EventBus.Port,
		}
		service.Meta["nats_request_prefix"] = cfg.Transport.NATS.ReqPrefix
	}

	if cfg.Transport.LoadBalancing.Enabled {
		service.Tags = append(service.Tags, "lb")
	}

	if Version != "" {
		service.Tags = append(service.Tags, Version)
	}

	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ctx.Done():
			err := client.Agent().ServiceDeregister(service.ID)
			if err != nil {
				log.Error(err.Error())
				return
			}

			log.Info("done")
			return

		case <-ticker.C:
			_, _, err := client.Agent().Service(service.ID, nil)
			if err != nil {
				err = client.Agent().ServiceRegister(service)
				if err != nil {
					log.Error(err.Error())
					continue
				}

				log.Info("service registration")
			}
		}
	}
}
