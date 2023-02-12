package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	router := gin.Default()
	router.GET("/users", getUsers)

	router.Run("localhost:8000")
}

func getUsers(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, users)
}
