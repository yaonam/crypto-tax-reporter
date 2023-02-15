package main

import (
	"encoding/json"
	"net/http"

	// "encoding/json"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type user struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

var users = []user{
	{ID: "1", FirstName: "Elim", LastName: "Poon", Email: "elimviolinist@gmail.com"},
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/users", getUsers)
	// r.Get("/user/:id", getUser)
	// r.Post("/user", postUser)

	http.ListenAndServe("127.0.0.1:8000", r)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	if res, err := json.Marshal(&users); err == nil {
		w.Write(res)
		// w.Write([]byte("This works?"))
	} else {
		panic("Get user request failed!" + err.Error())
	}
}

// func getUser(w http.ResponseWriter, r *http.Request) {
// 	id := c.Param("id")

// 	for _, user := range users {
// 		if user.ID == id {
// 			c.IndentedJSON(http.StatusOK, user)
// 			return
// 		}
// 	}
// 	c.IndentedJSON(http.StatusNotFound, {"message": "album not found"})
// }

// func postUser(w http.ResponseWriter, r *http.Request) {
// 	var newUser user

// 	// Bind received JSON to newUser.
// 	if err := c.BindJSON(&newUser); err != nil {
// 		return
// 	}

// 	// Add newUser to users.
// 	users = append(users, newUser)
// 	c.IndentedJSON(http.StatusCreated, newUser)
// }
