package nats

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/pubsub"
)

type natsTestSuite struct {
	suite.Suite
	cfg    conf.EventBus
	pubSub pubsub.PubSub
}

func (suite *natsTestSuite) SetupSuite() {
	path, ok := os.LookupEnv("IDENTITY_PATH")
	if !ok {
		path = "../.."
	}

	cfg, err := conf.LoadConfig(path)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	cfg.EventBus.Users = conf.Users{
		Stream: conf.Stream{
			Name: "TESTS",
			Config: []byte(`{
					"subjects": [
						"tests.>"
					],
					"retention": "interest",
					"storage": "memory"
				}`),
		},
		Consumer: conf.Consumer{
			Name:   "test-1",
			Stream: "TESTS",
			Config: []byte(`{}`),
		},
	}

	pubSub, err := NewPubSub(cfg.EventBus)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	pullBasedPubSub, _ := pubSub.PullBasedPubSub()

	stream := cfg.EventBus.Users.Stream
	if err := pullBasedPubSub.AddStream(stream.Name, stream.Config); err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.cfg = cfg.EventBus
	suite.pubSub = pubSub
}

func (suite *natsTestSuite) TestPublishAndSubscribe() {
	data := make(chan string, 1)

	err := suite.pubSub.Subscribe("tests.>", func(ctx context.Context, msg *pubsub.Message) error {
		data <- string(msg.Data)
		return nil
	})
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.pubSub.Publish("tests.hello", []byte("world"))

	ack := <-data
	suite.Equal("world", ack)
}

func (suite *natsTestSuite) TestPullSubscribe() {
	pullBasedPubSub, _ := suite.pubSub.PullBasedPubSub()

	stream := suite.cfg.Users.Stream
	consumer := suite.cfg.Users.Consumer
	if err := pullBasedPubSub.AddConsumer(consumer.Name, stream.Name, consumer.Config); err != nil {
		suite.Fail(err.Error())
		return
	}

	data := make(chan string, 1)
	if err := pullBasedPubSub.PullSubscribe(consumer.Name, stream.Name, func(ctx context.Context, msg *pubsub.Message) error {
		data <- string(msg.Data)
		return nil
	}); err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.pubSub.Publish("tests.hello", []byte("world"))

	ack := <-data
	suite.Equal("world", ack)
}

func (suite *natsTestSuite) TearDownSuite() {
	suite.pubSub.Close()
}

func TestNatsTestSuite(t *testing.T) {
	suite.Run(t, new(natsTestSuite))
}
