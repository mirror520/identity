package pubsub

import "encoding/json"

type PubSub interface {
	Publish(topic string, data []byte) error
	Subscribe(topic string, callback MessageHandler) error
	Close() error
}

type PullBasedPubSub interface {
	AddStream(name string, raw json.RawMessage) error
	AddConsumer(name string, stream string, raw json.RawMessage) error
	PullSubscribe(consumer string, stream string, callback MessageHandler) error
}

type MessageHandler func(msg *Message) error

type Message struct {
	Header map[string][]string
	Data   []byte
}
