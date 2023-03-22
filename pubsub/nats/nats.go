package nats

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/model"
	"github.com/mirror520/identity/pubsub"
)

type ConsumerStream struct {
	Consumer string
	Stream   string
}

func NewPubSub(cfg conf.EventBus) (pubsub.PubSub, error) {
	log := zap.L().With(
		zap.String("pubsub", "nats"),
	)

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

	ctx, cancel := context.WithCancel(context.Background())

	return &pubSub{
		log:           log,
		nc:            nc,
		js:            js,
		subscriptions: make(map[string]*nats.Subscription),
		cancels:       make(map[ConsumerStream]context.CancelFunc),
		rootCtx:       ctx,
		rootCancel:    cancel,
	}, nil
}

func NewPullBasedPubSub(cfg conf.EventBus) (pubsub.PullBasedPubSub, error) {
	pubSub, err := NewPubSub(cfg)
	if err != nil {
		return nil, err
	}
	return pubSub.PullBasedPubSub()
}

type pubSub struct {
	log           *zap.Logger
	nc            *nats.Conn
	js            nats.JetStreamContext
	subscriptions map[string]*nats.Subscription         // map[topic]*nats.Subscription
	cancels       map[ConsumerStream]context.CancelFunc // map[ConsumerStream]context.CancelFunc
	rootCtx       context.Context
	rootCancel    context.CancelFunc
	sync.Mutex
}

func (ps *pubSub) Publish(topic string, data []byte) error {
	_, err := ps.js.Publish(topic, data)
	return err
}

func (ps *pubSub) Subscribe(topic string, callback pubsub.MessageHandler) error {
	sub, err := ps.js.Subscribe(topic, func(m *nats.Msg) {
		msg := &pubsub.Message{
			Topic: m.Subject,
			Data:  m.Data,
		}
		callback(context.Background(), msg)
	})

	if err != nil {
		return err
	}

	ps.Lock()
	ps.subscriptions[topic] = sub
	ps.Unlock()
	return nil
}

func (ps *pubSub) PullBasedPubSub() (pubsub.PullBasedPubSub, error) {
	var pullBasedPubSub pubsub.PullBasedPubSub = ps
	return pullBasedPubSub, nil
}

func (ps *pubSub) AddStream(name string, raw json.RawMessage) error {
	var cfg *nats.StreamConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	cfg.Name = name

	_, err := ps.js.AddStream(cfg)
	return err
}

func (ps *pubSub) AddConsumer(name string, stream string, raw json.RawMessage) error {
	var cfg *nats.ConsumerConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	cfg.Durable = name

	_, err := ps.js.AddConsumer(stream, cfg)
	return err
}

func (ps *pubSub) PullSubscribe(consumer string, stream string, callback pubsub.MessageHandler) error {
	log := ps.log.With(
		zap.String("action", "pull_subscribe"),
		zap.String("consumer", consumer),
		zap.String("stream", stream),
	)

	sub, err := ps.js.PullSubscribe("", consumer, nats.BindStream(stream))
	if err != nil {
		return err
	}

	cs := ConsumerStream{
		Consumer: consumer,
		Stream:   stream,
	}

	ctx := context.WithValue(ps.rootCtx, model.LOGGER, log)
	ctx, cancel := context.WithCancel(ctx)
	ps.Lock()
	if cancel, ok := ps.cancels[cs]; ok {
		cancel()
	}
	ps.cancels[cs] = cancel
	ps.Unlock()

	go ps.pull(ctx, sub, callback)

	return nil
}

func (ps *pubSub) pull(ctx context.Context, sub *nats.Subscription, callback pubsub.MessageHandler) {
	log, ok := ctx.Value(model.LOGGER).(*zap.Logger)
	if !ok {
		log = ps.log
	}

	for {
		select {
		case <-ctx.Done():
			sub.Unsubscribe()
			log.Info("done")
			return

		default:
			msgs, err := sub.Fetch(100)
			if err != nil && !errors.Is(err, nats.ErrTimeout) {
				log.Error(err.Error())
				continue
			}

			for _, m := range msgs {
				msg := &pubsub.Message{
					Topic: m.Subject,
					Data:  m.Data,
				}

				err := callback(context.Background(), msg)
				if err != nil {
					meta, metaErr := m.Metadata()
					if metaErr != nil {
						log.Error(metaErr.Error(),
							zap.String("topic", m.Subject),
						)

						continue
					}

					log.Error(err.Error(),
						zap.String("topic", m.Subject),
						zap.Uint64("stream_seq", meta.Sequence.Stream),
						zap.Uint64("consumer_seq", meta.Sequence.Consumer),
					)
					continue
				}

				m.Ack()
			}
		}
	}
}

func (ps *pubSub) Close() error {
	ps.rootCancel()

	for _, sub := range ps.subscriptions {
		sub.Unsubscribe()
		sub.Drain()
	}

	return ps.nc.Drain()
}
