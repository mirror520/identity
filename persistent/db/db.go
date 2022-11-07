package db

import "gorm.io/gorm"

type DBPersistent interface {
	DB() *gorm.DB
}
