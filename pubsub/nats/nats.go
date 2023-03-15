package nats

import (
	"os"

	"github.com/nats-io/nats.go"

	"github.com/mirror520/identity/pubsub"
)

func NewPubSub() (pubsub.PubSub, error) {
	url, ok := os.LookupEnv("NATS_URL")
	if !ok {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	ps := new(pubSub)
	ps.nc = nc
	ps.js = js
	ps.subscriptions = make(map[string]*nats.Subscription)

	return ps, nil
}

type pubSub struct {
	nc            *nats.Conn
	js            nats.JetStreamContext
	subscriptions map[string]*nats.Subscription // map[topic]*nats.Subscription
}

func (ps *pubSub) Publish(topic string, data []byte) error {
	return ps.nc.Publish(topic, data)
}

func (ps *pubSub) Subscribe(topic string, callback pubsub.MessageHandler) error {
	sub, err := ps.nc.Subscribe(topic, func(m *nats.Msg) {
		msg := &pubsub.Message{
			Header: m.Header,
			Data:   m.Data,
		}
		callback(msg)
	})

	if err != nil {
		return err
	}

	ps.subscriptions[topic] = sub
	return nil
}

func (ps *pubSub) Close() error {
	for _, sub := range ps.subscriptions {
		sub.Unsubscribe()
		sub.Drain()
	}

	return ps.nc.Drain()
}
