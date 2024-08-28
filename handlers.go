package main

import (
	"chariot-assessment/pkg/id"
	"encoding/json"
	"fmt"
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
	return
}

func createAccount(w http.ResponseWriter, r *http.Request) {
	return
}

func deposit(w http.ResponseWriter, r *http.Request) {
	return
}

func withdraw(w http.ResponseWriter, r *http.Request) {
	return
}
