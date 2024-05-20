package db

import (
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const file string = "../../database/imbere.db"

func DbCon() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(file), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	return db
}
