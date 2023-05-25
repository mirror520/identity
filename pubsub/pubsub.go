package pubsub

import (
	"context"
	"encoding/json"
)

type PubSub interface {
	Publish(topic string, data []byte) error
	Subscribe(topic string, callback MessageHandler) error
	Close() error

	PullBasedPubSub() (PullBasedPubSub, error)
}

type PullBasedPubSub interface {
	PubSub
	AddStream(name string, raw json.RawMessage) error
	AddConsumer(name string, stream string, raw json.RawMessage) error
	PullSubscribe(consumer string, stream string, callback MessageHandler) error
}

type MessageHandler func(ctx context.Context, msg *Message) error

type MessageResponse func(data []byte) error

type Message struct {
	Topic    string
	Data     []byte
	Response MessageResponse
}
