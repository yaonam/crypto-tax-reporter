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
}

type Asset struct {
	gorm.Model
	Title  string `json:"title"`
	Symbol string `json:"symbol"`
}

type Transaction struct {
	gorm.Model
	Timestamp string  `json:"timestamp"`
	Type      string  `json:"type"`
	From      Account `json:"from" gorm:"many2many:from_account;ForeignKey:id;References:id"`
	To        Account `json:"to" gorm:"many2many:to_account;ForeignKey:id;References:id"`
	Asset     Asset   `json:"asset" gorm:"ForeignKey:id;References:id"`
	Quantity  float64 `json:"quantity"`
	Currency  Asset   `json:"currency" gorm:"ForeignKey:id;References:id"`
	SpotPrice float64 `json:"spot_price"`
	Subtotal  float64 `json:"subtotal"`
	Total     float64 `json:"total"`
	Fees      float64 `json:"fees"`
	Notes     string  `json:"notes"`
}

type TaxLot struct {
	gorm.Model
	Timestamp string `json:"timestamp"`
	// owner
	// asset
	Quantity float64 `json:"quantity"`
	// Currency         float64 `json:"currency"`
	CostBasis        float64 `json:"cost_basis"`
	QuantityRealized float64 `json:"quantity_realized"`
}

func migrateModels(db *gorm.DB) {
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Account{})
	db.AutoMigrate(&Transaction{})
}
