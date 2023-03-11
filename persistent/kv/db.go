package kv

import "github.com/dgraph-io/badger/v3"

type Database interface {
	DB() *badger.DB
	Close() error
}
