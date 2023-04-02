package main

import (
	"crypto-tax-reporter/cmd/server"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	server.RunServer()
}
