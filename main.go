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
	r.HandleFunc("/health", health)

	r.HandleFunc("/users", createUser).
		Methods("POST")

	r.HandleFunc("/accounts", createAccount).
		Methods("POST")

	r.HandleFunc("/accounts/{account_id}/balance", getAccountBalance).
		Methods("GET")

	r.HandleFunc("/accounts/{account_id}/withdraw", withdraw).
		Methods("POST")
	r.HandleFunc("/accounts/{account_id}/deposit", deposit).
		Methods("POST")

	r.HandleFunc("/accounts/{account_id}/transfer", transfer).
		Methods("POST")

	r.HandleFunc("/transactions", listTransactions).
		Methods("GET")

	fmt.Println("Service ready.")

	http.ListenAndServe(":8080", r)
}
