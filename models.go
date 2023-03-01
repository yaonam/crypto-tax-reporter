package main

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

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

	TxFroms []Transaction `gorm:"foreignKey:From" json:"-"`
	TxTos   []Transaction `gorm:"foreignKey:To" json:"-"`
	TaxLots []TaxLot      `json:"-"`
}

type Asset struct {
	gorm.Model
	Title  string `json:"title"`
	Symbol string `json:"symbol"`

	TxAssets     []Transaction `gorm:"foreignKey:Asset" json:"-"`
	TxCurrencies []Transaction `gorm:"foreignKey:Currency" json:"-"`

	TaxLotAssets     []TaxLot `gorm:"foreignKey:Asset" json:"-"`
	TaxLotCurrencies []TaxLot `gorm:"foreignKey:Currency" json:"-"`
}

type Transaction struct {
	gorm.Model
	Timestamp string   `json:"timestamp"`
	Type      string   `json:"type"`
	From      uint     `json:"from"`
	To        uint     `json:"to"`
	Asset     uint     `json:"asset"`
	Quantity  float64  `json:"quantity"`
	Currency  uint     `json:"currency"`
	SpotPrice float64  `json:"spot_price"`
	Subtotal  float64  `json:"subtotal"`
	Total     float64  `json:"total"`
	Fees      float64  `json:"fees"`
	TaxLots   []TaxLot `json:"tax_lots"`
	Notes     string   `json:"notes"`
}

type TaxLot struct {
	gorm.Model
	Timestamp        string  `json:"timestamp"`
	AccountID        uint    `json:"account_id"`
	TransactionID    uint    `json:"transaction_id"`
	Asset            uint    `json:"asset"`
	Quantity         float64 `json:"quantity"`
	Currency         uint    `json:"currency"`
	CostBasis        float64 `json:"cost_basis"`
	QuantityRealized float64 `json:"quantity_realized"`
}

func migrateModels(db *gorm.DB) {
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Account{})
	db.AutoMigrate(&Asset{})
	db.AutoMigrate(&Transaction{})
	db.AutoMigrate(&TaxLot{})

	// newUser := User{FirstName: "Brandon", LastName: "Lee", Email: "brandon@gmail.com"}
	// db.Create(&newUser)
	// newAccount := Account{UserID: 1}
	// db.Create(&newAccount)
	// newTx := Transaction{From: 1}
	// db.Create(&newTx)
	// newTaxLot := TaxLot{TransactionID: 1}
	// db.Create(&newTaxLot)
}

// TODO Make this into an interface generic
func findAssetOrCreate(currency string) uint {
	var asset Asset
	result := db.Where("symbol = ?", currency).First(&asset)

	// If not found, createt new asset
	if result.Error == gorm.ErrRecordNotFound {
		newAsset := Asset{Title: currency, Symbol: currency}
		db.Create(&newAsset)
		log.Println("Createtd new Asset: " + currency + fmt.Sprint(newAsset.ID))
		return newAsset.ID
	}
	return asset.ID
}
