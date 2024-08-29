package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
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

	r := mux.NewRouter()
	r.PathPrefix("/health").HandlerFunc(health)

	r.PathPrefix("/users").HandlerFunc(createUser).
		Methods("POST")

	r.PathPrefix("/accounts").HandlerFunc(createAccount).
		Methods("POST")

	r.PathPrefix("/accounts/{account_id}/withdraw").HandlerFunc(withdraw).
		Methods("POST")
	r.PathPrefix("/accounts/{account_id}/deposit").HandlerFunc(deposit).
		Methods("POST")

	r.PathPrefix("/accounts/{account_id}/transfer").HandlerFunc(transfer).
		Methods("POST")

	fmt.Println("Service ready.")

	http.ListenAndServe(":8080", r)
}
