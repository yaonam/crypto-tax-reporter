package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"crypto-tax-reporter/cmd/models"
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	db.Model(&models.User{}).Preload("Accounts").Find(&users)
	if res, err := json.Marshal(&users); err == nil {
		w.Header().Set("Content-Type", "application/json") // json header
		w.Write(res)
	} else {
		panic("Failed to jsonify users!" + err.Error())
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")

	var user models.User
	db.Find(&user, id)

	if res, err := json.Marshal(&user); err == nil {
		w.Header().Set("Content-Type", "application/json") // json header
		w.Write(res)
	} else {
		panic("Get user request failed!" + err.Error())
	}
}

func postUser(w http.ResponseWriter, r *http.Request) {
	var newUser models.User

	// Bind received JSON to newUser.
	if err := render.Bind(r, &newUser); err != nil {
		panic("Invalid request")
	}

	// Create
	db.Create(&newUser)

	w.Write([]byte(fmt.Sprintf("Post user %v %v successful!", newUser.ID, newUser.FirstName)))
}

func getAccounts(w http.ResponseWriter, r *http.Request) {
	var accounts []models.Account
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
	var newAccount models.Account

	// Bind received JSON to newUser.
	if err := render.Bind(r, &newAccount); err != nil {
		panic("Invalid request")
	}

	// Create
	db.Create(&newAccount)

	w.Write([]byte(fmt.Sprintf("Post account %v %v successful!", newAccount.ID, newAccount.UserID)))
}

func getTransactions(w http.ResponseWriter, r *http.Request) {
	var transactions []models.Transaction
	db.Model(&models.Transaction{}).Preload("TaxLots").Find(&transactions)
	if res, err := json.Marshal(&transactions); err == nil {
		w.Header().Set("Content-Type", "application/json") // json header
		w.Write(res)
	} else {
		panic("Failed to jsonify accounts!" + err.Error())
	}
}

func postTransaction(w http.ResponseWriter, r *http.Request) {
	var newTransaction models.Transaction

	// Bind received JSON to newUser.
	if err := render.Bind(r, &newTransaction); err != nil {
		panic("Invalid request")
	}

	// Create
	db.Create(&newTransaction)

	w.Write([]byte(fmt.Sprintf("Post account %v successful!", newTransaction)))
}
