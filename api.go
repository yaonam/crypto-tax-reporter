package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	db.Model(&User{}).Preload("Accounts").Find(&users)
	if res, err := json.Marshal(&users); err == nil {
		w.Header().Set("Content-Type", "application/json") // json header
		w.Write(res)
	} else {
		panic("Failed to jsonify users!" + err.Error())
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")

	var user User
	db.Find(&user, id)

	if res, err := json.Marshal(&user); err == nil {
		w.Header().Set("Content-Type", "application/json") // json header
		w.Write(res)
	} else {
		panic("Get user request failed!" + err.Error())
	}
}

func postUser(w http.ResponseWriter, r *http.Request) {
	var newUser User

	// Bind received JSON to newUser.
	if err := render.Bind(r, &newUser); err != nil {
		panic("Invalid request")
	}

	// Create
	db.Create(&newUser)

	w.Write([]byte(fmt.Sprintf("Post user %v %v successful!", newUser.ID, newUser.FirstName)))
}

func (u *User) Bind(r *http.Request) error {
	if u == nil {
		return errors.New("Missing user field")
	}
	return nil
}

func getAccounts(w http.ResponseWriter, r *http.Request) {
	var accounts []Account
	db.Find(&accounts)
	if res, err := json.Marshal(&accounts); err == nil {
		w.Header().Set("Content-Type", "application/json") // json header
		w.Write(res)
	} else {
		panic("Failed to jsonify accounts!" + err.Error())
	}
}

// func getAccount(w http.ResponseWriter, r *http.Request) {
// 	id := chi.URLParam(r, "userId")

// 	var user User
// 	db.Find(&user, id)

// 	if res, err := json.Marshal(&user); err == nil {
// 		w.Header().Set("Content-Type", "application/json") // json header
// 		w.Write(res)
// 	} else {
// 		panic("Get user request failed!" + err.Error())
// 	}
// }

func postAccount(w http.ResponseWriter, r *http.Request) {
	var newAccount Account

	// Bind received JSON to newUser.
	if err := render.Bind(r, &newAccount); err != nil {
		panic("Invalid request")
	}

	// Create
	db.Create(&newAccount)

	w.Write([]byte(fmt.Sprintf("Post account %v %v successful!", newAccount.ID, newAccount.UserID)))
}

func (u *Account) Bind(r *http.Request) error {
	if u == nil {
		return errors.New("Missing account field")
	}
	return nil
}

func getTransactions(w http.ResponseWriter, r *http.Request) {
	var transactions []Transaction
	db.Model(&Transaction{}).Preload("TaxLots").Find(&transactions)
	if res, err := json.Marshal(&transactions); err == nil {
		w.Header().Set("Content-Type", "application/json") // json header
		w.Write(res)
	} else {
		panic("Failed to jsonify accounts!" + err.Error())
	}
}

// func getAccount(w http.ResponseWriter, r *http.Request) {
// 	id := chi.URLParam(r, "userId")

// 	var user User
// 	db.Find(&user, id)

// 	if res, err := json.Marshal(&user); err == nil {
// 		w.Header().Set("Content-Type", "application/json") // json header
// 		w.Write(res)
// 	} else {
// 		panic("Get user request failed!" + err.Error())
// 	}
// }

func postTransaction(w http.ResponseWriter, r *http.Request) {
	var newTransaction Transaction

	// Bind received JSON to newUser.
	if err := render.Bind(r, &newTransaction); err != nil {
		panic("Invalid request")
	}

	// Create
	db.Create(&newTransaction)

	w.Write([]byte(fmt.Sprintf("Post account %v successful!", newTransaction)))
}

func (u *Transaction) Bind(r *http.Request) error {
	if u == nil {
		return errors.New("Missing account field")
	}
	return nil
}
