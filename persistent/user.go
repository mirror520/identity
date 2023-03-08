package persistent

import (
	"errors"

	"github.com/mirror520/identity/model/conf"
	"github.com/mirror520/identity/model/user"
	"github.com/mirror520/identity/persistent/db"
	"github.com/mirror520/identity/persistent/inmem"
	"github.com/mirror520/identity/persistent/kv"
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
