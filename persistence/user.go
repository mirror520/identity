package persistence

import (
	"errors"

	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/persistence/db"
	"github.com/mirror520/identity/persistence/inmem"
	"github.com/mirror520/identity/persistence/kv"
	"github.com/mirror520/identity/user"
)

func NewUserRepository(cfg conf.Persistence) (user.Repository, error) {
	switch cfg.Driver {
	case conf.SQLite:
		return db.NewUserRepository(cfg)
	case conf.BadgerDB:
		return kv.NewUserRepository(cfg)
	case conf.InMem:
		return inmem.NewUserRepository()
	default:
		return nil, errors.New("driver not supported")
	}
}
