package transport

import (
	"errors"
	"strconv"

	"github.com/go-kit/kit/sd"

	"github.com/mirror520/identity"
	"github.com/mirror520/identity/transport/pubsub"
)

var (
	signInFactories = make(map[string]sd.Factory)
)

func MakeEndpoints(instance identity.Instance) (*identity.EndpointSet, error) {
	endpoints := new(identity.EndpointSet)
	empty := true

	switch instance.Protocol {
	case "nats":
		url := instance.Address + ":" + strconv.Itoa(instance.Port)
		factory, ok := signInFactories[url]
		if !ok {
			f, err := pubsub.SignInFactory(instance.Address, instance.Port)
			if err != nil {
				return nil, err
			}

			factory = f
			signInFactories[url] = factory
		}

		signIn, _, err := factory(instance.RequestPrefix)
		if err != nil {
			return nil, err
		}

		endpoints.SignIn = signIn
		empty = false
	}

	if empty {
		return nil, errors.New("endpoint empty")
	}

	return endpoints, nil
}
