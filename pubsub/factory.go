package pubsub

import (
	"errors"

	"github.com/mirror520/identity/conf"
)

type factory func(cfg conf.Instance) (PubSub, error)

var factories = make(map[conf.TransportProvider]factory)

func AddFactory(provider conf.TransportProvider, factory factory) {
	factories[provider] = factory
}

func NewPubSub(provider conf.TransportProvider, cfg conf.Instance) (PubSub, error) {
	factory, ok := factories[provider]
	if !ok {
		return nil, errors.New("provider not supported")
	}

	return factory(cfg)
}
