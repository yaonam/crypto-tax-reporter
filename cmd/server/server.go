package server

import (
	"log"
	"os"

	"github.com/glebarez/sqlite" // Pure go driver, doesn't need cgo
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"crypto-tax-reporter/cmd/models"
	"crypto-tax-reporter/cmd/wallet"
)

var db *gorm.DB

func RunServer() {
	log.SetOutput(os.Stdout)
	os.Remove("test.db")
	var err error
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	models.MigrateModels(db)

	// Dev, temp func calls
	wallet.Import(db, 1, "0x844e94FC29D39840229F6E47290CbE73f187c3b1")

	// coinbase.OpenFile(db, 1, "csv/data.csv")
	// pnl := taxes.CalculateUserPNL(db, 1)
	// log.Printf("PNL: %v", pnl)

	// r := chi.NewRouter()
	// r.Use(middleware.Logger)

	// r.Get("/", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("Welcome!")) })

	// r.Route("/user", func(r chi.Router) {
	// 	r.Get("/", getUsers)
	// 	r.Post("/", postUser)
	// 	r.Get("/{userId}", getUser)
	// })
	// r.Route("/account", func(r chi.Router) {
	// 	r.Get("/", getAccounts)
	// 	r.Post("/", postAccount)
	// 	// r.Get("/{userId}", getUser)
	// })
	// r.Route("/transaction", func(r chi.Router) {
	// 	r.Get("/", getTransactions)
	// 	r.Post("/", postTransaction)
	// 	// r.Get("/{userId}", getUser)
	// })

	log.Println("Server started")
	// http.ListenAndServe("127.0.0.1:8000", r)
}
