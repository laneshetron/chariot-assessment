package main

import (
	"errors"
)

func CreateSchema() error {
	if pgClient == nil {
		return errors.New("postgres client has not been initialized.")
	}
	// balances use a decimal precision of 15,4
	// for forward-compatibility w/ other currencies
	_, err := pgClient.Exec(`
        DO $$ BEGIN
            CREATE TYPE t_transaction AS ENUM
                ('withdrawal', 'deposit', 'transfer_in', 'transfer_out');
        EXCEPTION
            WHEN duplicate_object THEN null;
        END $$;

        CREATE TABLE IF NOT EXISTS users(
            id varchar(20) PRIMARY KEY,
            name text,
            created_at timestamp DEFAULT current_timestamp
        );
        CREATE TABLE IF NOT EXISTS accounts(
            id varchar(20) PRIMARY KEY,
            user_id varchar(20) NOT NULL,
            balance decimal(15,4) NOT NULL DEFAULT 0.0,
            created_at timestamp DEFAULT current_timestamp,
            FOREIGN KEY (user_id) REFERENCES users(id)
        );
        CREATE TABLE IF NOT EXISTS transactions(
            id varchar(20) PRIMARY KEY,
            related_transaction_id varchar(20),
            account_id varchar(20) NOT NULL,
            external_account varchar(20),
            idempotency_key varchar(100) NOT NULL,
            amount decimal(15,4) NOT NULL,
            ending_balance decimal(15,4) NOT NULL,
            type t_transaction,
            created_at timestamp DEFAULT current_timestamp,
            FOREIGN KEY (account_id) REFERENCES accounts(id),
            FOREIGN KEY (external_account) REFERENCES accounts(id),
            FOREIGN KEY (related_transaction_id) REFERENCES transactions(id),
            UNIQUE (idempotency_key, type)
        );
    `)

	return err
}
