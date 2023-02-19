package main

import (
	"net/http"
)

func test(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome!"))
}
