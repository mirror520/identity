package db

import "gorm.io/gorm"

type Database interface {
	DB() *gorm.DB
}
