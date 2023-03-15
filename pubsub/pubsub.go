package pubsub

type PubSub interface {
	Publish(topic string, data []byte) error
	Subscribe(topic string, callback MessageHandler) error
	Close() error
}

type MessageHandler func(msg *Message)

type Message struct {
	Header map[string][]string
	Data   []byte
}
