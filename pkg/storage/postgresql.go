package storage

import (
	"database/sql"
	"time"
)

var db *sql.DB

func initPostgresql(dataSourceName string) {
	var err error
	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	db.SetConnMaxLifetime(time.Hour)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(200)
}

func GetConnectionPool(dataSourceName string) *sql.DB {
	if db == nil {
		initPostgresql(dataSourceName)
	}
	return db
}
