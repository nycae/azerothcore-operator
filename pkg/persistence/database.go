package persistence

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbUser = getEnv("DB_USER", "root")
	dbPass = getEnv("DB_PASS", "FvfvgNx9d1xsJIJSZaH7Fqws")
	dbHost = getEnv("DB_HOST", "localhost")
	dbPort = getEnv("DB_PORT", "3306")
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func DefaultDatabaseConnectionString(database string) string {
	return DatabaseConnectionString(dbUser, dbPass, dbHost, dbPort, database)
}

func DatabaseConnectionString(user, pass, host, port, database string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, host, port, database)
}

func AuthDB() *sql.DB {
	db, err := sql.Open("mysql", DefaultDatabaseConnectionString("acore_auth"))
	if err != nil {
		log.Fatalf("unable to open database: %v", err)
	}
	return db
}
