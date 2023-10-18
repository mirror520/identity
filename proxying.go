package identity

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/go-kit/kit/endpoint"

	"github.com/mirror520/identity/user"
)

type Instance struct {
	ID            string
	Protocol      string
	Address       string
	Port          int
	RequestPrefix string
	ModifiedTime  time.Time
	IsAlive       bool

	Endpoints *EndpointSet
}

func ProxyingMiddleware(ctx context.Context, ch <-chan Instance) ServiceMiddleware {
	return func(next Service) Service {
		mw := &proxyingMiddleware{
			instances: make([]*Instance, 0),
			next:      next,
		}

		go mw.updateInstances(ctx, ch)

		return mw
	}
}

type proxyingMiddleware struct {
	next      Service
	instances []*Instance
	n         int
	sync.RWMutex
}

func (mw *proxyingMiddleware) updateInstances(ctx context.Context, ch <-chan Instance) {
	for {
		select {
		case <-ctx.Done():
			return

		case new := <-ch:
			mw.Lock()

			processed := false
			for i, instance := range mw.instances {
				if instance.ID != new.ID {
					continue
				}

				if !instance.IsAlive {
					mw.instances = append(mw.instances[:i], mw.instances[i+1:]...)
				} else {
					if new.ModifiedTime.After(instance.ModifiedTime) {
						mw.instances[i] = &new
					}
				}

				processed = true
				break
			}

			if !processed {
				mw.instances = append(mw.instances, &new)
			}

			mw.Unlock()
		}
	}
}

func (mw *proxyingMiddleware) Endpoint(method string) (endpoint.Endpoint, bool) {
	mw.RLock()
	defer mw.RUnlock()

	size := len(mw.instances)
	if size == 0 {
		return nil, false
	}

	instance := mw.instances[mw.n%size] // lb: rr
	mw.n++

	switch method {
	case "SignIn":
		if instance.Endpoints.SignIn == nil {
			return nil, false
		}
		return instance.Endpoints.SignIn, true

	default:
		return nil, false
	}
}

func (mw *proxyingMiddleware) Register(username string, name string, email string) (*user.User, error) {
	return mw.next.Register(username, name, email)
}

func (mw *proxyingMiddleware) OTPVerify(otp string, id user.UserID) (*user.User, error) {
	return mw.next.OTPVerify(otp, id)
}

func (mw *proxyingMiddleware) SignIn(credential string, provider user.SocialProvider) (*user.User, error) {
	endpoint, ok := mw.Endpoint("SignIn")
	if !ok {
		return mw.next.SignIn(credential, provider)
	}

	req := &SignInRequest{
		Credential: credential,
		Provider:   provider,
	}

	resp, err := endpoint(context.Background(), req)
	if err != nil {
		return nil, err
	}

	u, ok := resp.(*user.User)
	if !ok {
		return nil, errors.New("invalid user")
	}

	return u, nil
}

func (mw *proxyingMiddleware) AddSocialAccount(credential string, provider user.SocialProvider, id user.UserID) (*user.User, error) {
	return mw.next.AddSocialAccount(credential, provider, id)
}

func (mw *proxyingMiddleware) CheckHealth(ctx context.Context) error {
	return mw.next.CheckHealth(ctx)
}

func (mw *proxyingMiddleware) Handler() (EventHandler, error) {
	return mw.next.Handler()
}
