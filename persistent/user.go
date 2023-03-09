package persistent

import (
	"errors"

	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/persistent/db"
	"github.com/mirror520/identity/persistent/inmem"
	"github.com/mirror520/identity/persistent/kv"
	"github.com/mirror520/identity/user"
)

func NewUserRepository(cfg conf.DB) (user.Repository, error) {
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
