package nats

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/pubsub"
)

type natsTestSuite struct {
	suite.Suite
	pubSub pubsub.PubSub
}

func (suite *natsTestSuite) SetupSuite() {
	cfg := &conf.Config{
		Name: "identity-1",
	}

	pubSub, err := NewPubSub(cfg)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	jsonStr := `{
		"subjects": [ 
			"tests.>" 
		],
		"retention": "interest",
		"storage": "memory"
	}`

	pullBasedPubSub := pubSub.(pubsub.PullBasedPubSub)
	if err := pullBasedPubSub.AddStream("TESTS", []byte(jsonStr)); err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.pubSub = pubSub
}

func (suite *natsTestSuite) TestPublishAndSubscribe() {
	data := make(chan string, 1)

	err := suite.pubSub.Subscribe("tests.>", func(msg *pubsub.Message) error {
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
	pullBasedPubSub := suite.pubSub.(pubsub.PullBasedPubSub)
	if err := pullBasedPubSub.AddConsumer("instance-1", "TESTS", []byte(`{}`)); err != nil {
		suite.Fail(err.Error())
		return
	}

	data := make(chan string, 1)

	if err := pullBasedPubSub.PullSubscribe("instance-1", "TESTS", func(msg *pubsub.Message) error {

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
