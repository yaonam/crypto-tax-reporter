# crypto-tax-reporter

## Overview

A crypto tax reporting service built on Go and Gin.

## Notes

Modules -> Packages -> Files

`nodemon --watch './**/*.go' --signal SIGTERM --exec 'go' run main.go`

`go mod init {MODULE_NAME}` to create project dependency file
`go get .` to automatically track dependencies (based on imports)
`go run .` to run server
`go mode tidy` to remove unused dependencies
