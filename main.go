package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/glebarez/sqlite" // Pure go driver, doesn't need cgo
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	log.SetOutput(os.Stdout)
	var err error
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	migrateModels(db)

	// openFile()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("Welcome!")) })

	r.Route("/user", func(r chi.Router) {
		r.Get("/", getUsers)
		r.Post("/", postUser)
		r.Get("/{userId}", getUser)
	})

	log.Println("Server started")
	http.ListenAndServe("127.0.0.1:8000", r)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	db.Find(&users)
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
