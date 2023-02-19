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

	// "gorm.io/driver/sqlite"
	"github.com/glebarez/sqlite" // Pure go, doesn't need cgo
	"gorm.io/gorm"
)

type UserModel struct {
	gorm.Model
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

var db *gorm.DB

func main() {
	os.Setenv("CGO_ENABLED", "1")

	log.SetOutput(os.Stdout)
	var err error
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&UserModel{})

	// // Create
	// db.Create(&UserModel{FirstName: "Elim", LastName: "Poon"})

	// // Read
	// var product UserModel
	// db.First(&product, 1) // find product with integer primary key
	// db.First(&product, "code = ?", "D42") // find product with code D42

	// // Update - update product's price to 200
	// db.Model(&product).Update("Price", 200)
	// // Update - update multiple fields
	// db.Model(&product).Updates(UserModel{Price: 200, Code: "F42"}) // non-zero fields
	// db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// // Delete - delete product
	// db.Delete(&product, 1)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/users", getUsers)
	r.Route("/user", func(r chi.Router) {
		r.Post("/", postUser)
		r.Get("/{userId}", getUser)
	})

	log.Println("Server started!")
	http.ListenAndServe("127.0.0.1:8000", r)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []UserModel
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

	var user UserModel
	db.Find(&user, id)

	if res, err := json.Marshal(&user); err == nil {
		w.Header().Set("Content-Type", "application/json") // json header
		w.Write(res)
	} else {
		panic("Get user request failed!" + err.Error())
	}
}

func postUser(w http.ResponseWriter, r *http.Request) {
	var newUser UserModel

	// Bind received JSON to newUser.
	if err := render.Bind(r, &newUser); err != nil {
		panic("Invalid request")
	}

	// Create
	db.Create(&newUser)

	w.Write([]byte(fmt.Sprintf("Post user %v %v successful!", newUser.ID, newUser.FirstName)))
}

func (u *UserModel) Bind(r *http.Request) error {
	if u == nil {
		return errors.New("Missing user field")
	}
	return nil
}
