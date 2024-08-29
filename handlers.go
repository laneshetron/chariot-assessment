package main

import (
	"chariot-assessment/pkg/id"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"io"
	"io/ioutil"
	"net/http"
)

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type NewUser struct {
	Name string `json:"name"`
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user NewUser

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Could not read body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer func() {
		// Drain the request body
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}()

	err = json.Unmarshal(body, &user)
	if err != nil {
		fmt.Println("Could not unmarshal request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId, err := id.New()
	if err != nil {
		fmt.Println("Could not generate ID:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = pgClient.Exec(`
    INSERT INTO users(id, name) VALUES ($1, $2)
    `, userId, user.Name)
	if err != nil {
		fmt.Println("Error while inserting into postgres:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

type NewAccount struct {
	UserId string `json:"userId"`
}

func createAccount(w http.ResponseWriter, r *http.Request) {
	// XXX as with all of these endpoints, in a real service we would pull
	// userId from a signed JWT or some other authentication source, but for
	// the purposes of this assignment we'll assume the client is telling the truth.
	var accountReq NewAccount

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Could not read body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer func() {
		// Drain the request body
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}()

	err = json.Unmarshal(body, &accountReq)
	if err != nil {
		fmt.Println("Could not unmarshal request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accountId, err := id.New()
	if err != nil {
		fmt.Println("Could not generate ID:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = pgClient.Exec(`
    INSERT INTO accounts(id, user_id) VALUES ($1, $2)
    `, accountId, accountReq.UserId)
	if err != nil {
		fmt.Println("Error while inserting into postgres:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

type DepositWithdrawRequest struct {
	Amount         float64 `json:"amount"`
	IdempotencyKey string  `json:"idempotencyKey"`
}

func deposit(w http.ResponseWriter, r *http.Request) {
	var req DepositWithdrawRequest

	vars := mux.Vars(r)
	accountId := vars["account_id"]
	if accountId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Could not read body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer func() {
		// Drain the request body
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}()

	err = json.Unmarshal(body, &req)
	if err != nil {
		fmt.Println("Could not unmarshal request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Start a transaction with serializable isolation level
	tx, err := pgClient.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		fmt.Println("Could not begin transaction:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var newBalance float64
	err = tx.QueryRow(`
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2
		FOR UPDATE
		RETURNING balance
	`, req.Amount, accountId).Scan(&newBalance)
	if err != nil {
		fmt.Println("Error while updating account balance:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Generate transaction ID
	transactionId, err := id.New()
	if err != nil {
		fmt.Println("Could not generate ID:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Insert transaction record with ending_balance and idempotency_key
	_, err = tx.Exec(`
		INSERT INTO transactions(id, account_id, amount, type, ending_balance, idempotency_key)
		VALUES ($1, $2, $3, 'deposit', $4, $5)
	`, transactionId, accountId, req.Amount, newBalance, req.IdempotencyKey)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			// Unique violation error code, idempotency key already exists
			w.WriteHeader(http.StatusOK)
			return
		}
		fmt.Println("Error while inserting transaction:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		fmt.Println("Error while committing transaction:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func withdraw(w http.ResponseWriter, r *http.Request) {
	var req DepositWithdrawRequest

	vars := mux.Vars(r)
	accountId := vars["account_id"]
	if accountId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Could not read body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer func() {
		// Drain the request body
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}()

	err = json.Unmarshal(body, &req)
	if err != nil {
		fmt.Println("Could not unmarshal request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Start a transaction with serializable isolation level
	tx, err := pgClient.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		fmt.Println("Could not begin transaction:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Check for sufficient balance & update it
	var newBalance float64
	err = tx.QueryRow(`
		UPDATE accounts
		SET balance = balance - $1
		WHERE id = $2 AND balance >= $1
		FOR UPDATE
		RETURNING balance
	`, req.Amount, accountId).Scan(&newBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Could not withdraw: Insufficient funds")
			w.WriteHeader(http.StatusBadRequest)
		} else {
			fmt.Println("Error while updating account balance:", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Generate transaction ID
	transactionId, err := id.New()
	if err != nil {
		fmt.Println("Could not generate ID:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Insert transaction record with ending_balance and idempotency_key
	_, err = tx.Exec(`
		INSERT INTO transactions(id, account_id, amount, type, ending_balance, idempotency_key)
		VALUES ($1, $2, $3, 'withdraw', $4, $5)
	`, transactionId, accountId, req.Amount, newBalance, req.IdempotencyKey)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			// Unique violation error code, idempotency key already exists
			w.WriteHeader(http.StatusOK)
			return
		}
		fmt.Println("Error while inserting transaction:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		fmt.Println("Error while committing transaction:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type TransferRequest struct {
	Amount          float64 `json:"amount"`
	IdempotencyKey  string  `json:"idempotencyKey"`
	ExternalAccount string  `json:"externalAccount,omitempty"`
}

func transfer(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest

	vars := mux.Vars(r)
	accountId := vars["account_id"]
	if accountId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Could not read body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer func() {
		// Drain the request body
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}()

	err = json.Unmarshal(body, &req)
	if err != nil {
		fmt.Println("Could not unmarshal request body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Start a transaction with serializable isolation level
	tx, err := pgClient.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		fmt.Println("Could not begin transaction:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Check for sufficient balance & update sender's balance
	var senderNewBalance float64
	err = tx.QueryRow(`
		UPDATE accounts
		SET balance = balance - $1
		WHERE id = $2 AND balance >= $1
		FOR UPDATE
		RETURNING balance
	`, req.Amount, accountId).Scan(&senderNewBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Could not transfer: Insufficient funds")
			w.WriteHeader(http.StatusBadRequest)
		} else {
			fmt.Println("Error while updating sender's account balance:", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Update receiver's balance
	var receiverNewBalance float64
	err = tx.QueryRow(`
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2
		FOR UPDATE
		RETURNING balance
	`, req.Amount, req.ExternalAccount).Scan(&receiverNewBalance)
	if err != nil {
		fmt.Println("Error while updating receiver's account balance:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Generate transaction IDs
	senderTransactionId, err := id.New()
	if err != nil {
		fmt.Println("Could not generate sender transaction ID:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	receiverTransactionId, err := id.New()
	if err != nil {
		fmt.Println("Could not generate receiver transaction ID:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Insert sender's transaction record
	_, err = tx.Exec(`
		INSERT INTO transactions(id, account_id, amount, type, ending_balance,
            idempotency_key, related_transaction_id)
		VALUES ($1, $2, $3, 'transfer_out', $4, $5, $6)
	`, senderTransactionId, accountId, req.Amount, senderNewBalance,
		req.IdempotencyKey, receiverTransactionId)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			// Unique violation error code, idempotency key already exists
			w.WriteHeader(http.StatusOK)
			return
		}
		fmt.Println("Error while inserting sender's transaction:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Insert receiver's transaction record
	_, err = tx.Exec(`
		INSERT INTO transactions(id, account_id, amount, type, ending_balance,
            idempotency_key, related_transaction_id)
		VALUES ($1, $2, $3, 'transfer_in', $4, $5, $6)
	`, receiverTransactionId, req.ExternalAccount, req.Amount, receiverNewBalance,
		req.IdempotencyKey, senderTransactionId)
	if err != nil {
		fmt.Println("Error while inserting receiver's transaction:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		fmt.Println("Error while committing transaction:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
