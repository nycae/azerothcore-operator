package persistence

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbUser = getEnv("DB_USER", "")
	dbPass = getEnv("DB_PASS", "")
	dbHost = getEnv("DB_HOST", "localhost:3306")
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func AuthDB() *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/acore_auth", dbUser, dbPass, dbHost))
	if err != nil {
		log.Fatal(err)
	}
	return db
}
