package pubsub

import (
	"context"
)

// topic wildcards:
// * (star) can substitute for exactly one word.
// # (hash) can substitute for zero or more words.

type PubSub interface {
	Publish(topic string, data []byte) error
	Subscribe(topic string, callback MessageHandler) error
	Close() error
}

type MessageHandler func(ctx context.Context, msg *Message) error

type MessageResponse func(data []byte) error

type Message struct {
	Topic    string
	Data     []byte
	Response MessageResponse
}
