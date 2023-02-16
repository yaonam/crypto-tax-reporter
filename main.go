package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type User struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

var users = []User{
	{ID: "1", FirstName: "Elim", LastName: "Poon", Email: "elimviolinist@gmail.com"},
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/users", getUsers)
	r.Route("/user", func(r chi.Router) {
		r.Post("/", postUser)
		r.Get("/{userId}", getUser)
	})

	http.ListenAndServe("127.0.0.1:8000", r)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	if res, err := json.Marshal(&users); err == nil {
		w.Write(res)
	} else {
		panic("Get users request failed!" + err.Error())
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")

	for _, user := range users {
		if user.ID == id {
			if res, err := json.Marshal(&user); err == nil {
				w.Write(res)
			} else {
				panic("Get user request failed!" + err.Error())
			}
			return
		}
	}
	w.Write([]byte("No users"))
}

func postUser(w http.ResponseWriter, r *http.Request) {
	var newUser User

	// Bind received JSON to newUser.
	if err := render.Bind(r, &newUser); err != nil {
		panic("Invalid request")
	}

	// Add newUser to users.
	users = append(users, newUser)
	w.Write([]byte(fmt.Sprintf("Post user %v successful!", &newUser.FirstName)))
}

func (u *User) Bind(r *http.Request) error {
	if u == nil {
		return errors.New("Missing user field")
	}
	return nil
}
