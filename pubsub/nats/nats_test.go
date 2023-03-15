package nats

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mirror520/identity/pubsub"
)

type natsTestSuite struct {
	suite.Suite
	pubSub pubsub.PubSub
}

func (suite *natsTestSuite) SetupSuite() {
	pubSub, err := NewPubSub()
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.pubSub = pubSub
}

func (suite *natsTestSuite) TestPublishAndSubscribe() {
	data := make(chan string, 1)

	if err := suite.pubSub.Subscribe("ORDERS.>", func(msg *pubsub.Message) {
		data <- string(msg.Data)
	}); err != nil {
		suite.Fail(err.Error())
		return
	}

	err := suite.pubSub.Publish("ORDERS.scratch", []byte("hello"))
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	ack := <-data
	suite.Equal("hello", ack)
}

func (suite *natsTestSuite) TearDownSuite() {
	suite.pubSub.Close()
}

func TestNatsTestSuite(t *testing.T) {
	suite.Run(t, new(natsTestSuite))
}
