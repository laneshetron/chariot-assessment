package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var (
	pgClient *sql.DB
)

func main() {
	var err error
	// Setup postgres client
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, 5432, dbUser, dbPassword, dbName,
	)
	pgClient, err = sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	err = pgClient.Ping()
	if err != nil {
		panic(err)
	}

	// Setup database
	err = CreateSchema()
	if err != nil {
		panic(err)
	}

	fmt.Println("Service ready.")
}
