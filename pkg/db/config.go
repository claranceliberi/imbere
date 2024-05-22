package db

import (
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const file string = "./database/imbere.db"

func dbCon() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(file), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	return db
}

// Check database connection
// and Create tables in db;
func DbInit() {
	db := dbCon()

	db.AutoMigrate(&PullRequest{})
}
