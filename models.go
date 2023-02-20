package main

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Accounts  []Account `json:"accounts"`
}

type Account struct {
	gorm.Model
	UserID     uint   `json:"user_id"`
	Type       string `json:"type"`
	ExternalID string `json:"external_id"`

	FromTxs []Transaction `gorm:"foreignKey:From" json:"-"`
	ToTxs   []Transaction `gorm:"foreignKey:To" json:"-"`
}

type Asset struct {
	gorm.Model
	Title  string `json:"title"`
	Symbol string `json:"symbol"`
}

type Transaction struct {
	gorm.Model
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	From      uint   `json:"from"`
	To        uint   `json:"to"`
	// Asset     Asset    `json:"asset" gorm:"ForeignKey:id;References:id"`
	// Quantity  float64  `json:"quantity"`
	// Currency  Asset    `json:"currency" gorm:"ForeignKey:id;References:id"`
	// SpotPrice float64  `json:"spot_price"`
	// Subtotal  float64  `json:"subtotal"`
	// Total     float64  `json:"total"`
	// Fees      float64  `json:"fees"`
	// TaxLots   []TaxLot `json:"tax_lots"`
	// Notes     string   `json:"notes"`
}

type TaxLot struct {
	gorm.Model
	Timestamp        string  `json:"timestamp"`
	Account          Account `json:"account" gorm:"ForeignKey:id;References:id"`
	TransactionID    uint    `json:"transaction_id"`
	Asset            Asset   `json:"asset" gorm:"ForeignKey:id;References:id"`
	Quantity         float64 `json:"quantity"`
	Currency         Asset   `json:"currency" gorm:"ForeignKey:id;References:id"`
	CostBasis        float64 `json:"cost_basis"`
	QuantityRealized float64 `json:"quantity_realized"`
}

func migrateModels(db *gorm.DB) {
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Account{})
	db.AutoMigrate(&Asset{})
	db.AutoMigrate(&Transaction{})
	db.AutoMigrate(&TaxLot{})

	newUser := User{FirstName: "Brandon", LastName: "Lee", Email: "brandon@gmail.com"}
	db.Create(&newUser)
	newAccount := Account{UserID: 1}
	db.Create(&newAccount)
	newTx := Transaction{From: 1}
	db.Create(&newTx)
}
