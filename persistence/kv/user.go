package kv

import (
	"encoding/json"
	"errors"

	"github.com/dgraph-io/badger/v4"

	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/events"
	"github.com/mirror520/identity/user"
)

type userRepository struct {
	db *badger.DB
}

func NewUserRepository(cfg conf.Persistence) (user.Repository, error) {
	opts := badger.DefaultOptions(cfg.Host + "/" + cfg.Name)
	if cfg.InMem {
		opts = badger.DefaultOptions("").WithInMemory(true)
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	repo := new(userRepository)
	repo.db = db

	return repo, nil
}

func (repo *userRepository) Store(u *user.User) error {
	newUser := new(user.User)
	*newUser = *u

	u = newUser
	u.EventStore = nil

	bs, err := json.Marshal(u)
	if err != nil {
		return err
	}

	return repo.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(u.ID.Bytes(), bs)
		if err != nil {
			return err
		}

		err = txn.Set([]byte("username:"+u.Username), bs)
		if err != nil {
			return err
		}

		for _, account := range u.Accounts {
			err := txn.Set([]byte("social:"+account.SocialID), bs)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (repo *userRepository) Find(id user.UserID) (*user.User, error) {
	return repo.find(id.Bytes())
}

func (repo *userRepository) FindByUsername(username string) (*user.User, error) {
	return repo.find([]byte("username:" + username))
}

func (repo *userRepository) FindBySocialID(socialID user.SocialID) (*user.User, error) {
	return repo.find([]byte("social:" + socialID))
}

func (repo *userRepository) find(key []byte) (*user.User, error) {
	var u *user.User

	if err := repo.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return user.ErrUserNotFound
			}

			return err
		}

		return item.Value(func(val []byte) error {
			err := json.Unmarshal(val, &u)
			if err != nil {
				return err
			}

			u.EventStore = events.NewEventStore()
			return nil
		})
	}); err != nil {
		return nil, err
	}

	return u, nil
}

func (repo *userRepository) DB() *badger.DB {
	return repo.db
}

func (repo *userRepository) Close() error {
	return repo.db.Close()
}
