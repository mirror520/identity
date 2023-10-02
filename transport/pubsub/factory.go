package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/nats-io/nats.go"

	"github.com/mirror520/identity"
	"github.com/mirror520/identity/model"
	"github.com/mirror520/identity/user"
)

func SignInFactory(address string, port int) (sd.Factory, error) {
	url := "nats://" + address + ":" + strconv.Itoa(port)
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		return SignInEndpoint(nc, instance+".signin"), nil, nil
	}, nil
}

func SignInEndpoint(nc *nats.Conn, topic string) endpoint.Endpoint {
	return func(ctx context.Context, request any) (response any, err error) {
		req, ok := request.(*identity.SignInRequest)
		if !ok {
			return nil, errors.New("invalid request")
		}

		data, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}

		msg, err := nc.Request(topic, data, 5000*time.Millisecond)
		if err != nil {
			return nil, err
		}

		var result *model.Result
		if err := json.Unmarshal(msg.Data, &result); err != nil {
			return nil, err
		}

		if err := result.Error(); err != nil {
			return nil, err
		}

		var u *user.User
		if err := json.Unmarshal(result.Raw, &u); err != nil {
			return nil, err
		}

		return u, nil
	}
}
