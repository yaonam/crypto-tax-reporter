package models

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"gorm.io/gorm"
)

// TODO Change timestamps to time.Time type

type User struct {
	gorm.Model
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Accounts  []Account `gorm:"foreignKey:User" json:"accounts"`
}

type Account struct {
	gorm.Model
	User       uint   `json:"user"`
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
	From      uint     `json:"from"` // Account
	To        uint     `json:"to"`   // Account
	Asset     uint     `json:"asset"`
	Quantity  float64  `json:"quantity"`
	Currency  uint     `json:"currency"`
	SpotPrice float64  `json:"spot_price"`
	Subtotal  float64  `json:"subtotal"`
	Total     float64  `json:"total"`
	Fees      float64  `json:"fees"`
	TaxLots   []TaxLot `json:"tax_lots"`
	Notes     string   `json:"notes"`

	TaxLotSales []TaxLotSale `json:"-"`
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
	IsSold           bool    `json:"is_sold"`
	QuantityRealized float64 `json:"quantity_realized"`

	TaxLotSales []TaxLotSale `json:"-"`
}

type TaxLotSale struct {
	gorm.Model
	TransactionID uint        `json:"transaction_id"`
	Transaction   Transaction `json:"transaction"`
	TaxLotID      uint        `json:"taxlot_id"`
	TaxLot        TaxLot      `json:"taxlot"`
	QuantitySold  float64     `json:"quantity"`
}

func MigrateModels(db *gorm.DB) {
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Account{})
	db.AutoMigrate(&Asset{})
	db.AutoMigrate(&Transaction{})
	db.AutoMigrate(&TaxLot{})
	db.AutoMigrate(&TaxLotSale{})

	newUser := User{FirstName: "Elim", LastName: "Poon", Email: "elim@gmail.com"}
	db.FirstOrCreate(&newUser)
	newAccount := Account{User: newUser.ID}
	db.FirstOrCreate(&newAccount)
	newTx := Transaction{From: newAccount.ID}
	db.FirstOrCreate(&newTx)
	newTaxLot := TaxLot{TransactionID: 1}
	db.FirstOrCreate(&newTaxLot)
}

// TODO Make this into an interface generic
func FindAssetOrCreate(db *gorm.DB, currency string) uint {
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

// Finds or creates the account with externalID and then assigns to user
func AssignAccountToUser(db *gorm.DB, userID uint, externalID string) {
	accountID := FindAccountOrCreate(db, externalID)
	var account Account
	db.First(&account, accountID)
	account.User = userID
	db.Save(&account)
}

// Looks for the account under the user
func FindAccountOrCreate(db *gorm.DB, externalID string) uint {
	account := Account{ExternalID: externalID}
	db.FirstOrCreate(&account, account)
	return account.ID
}

func (u *User) Bind(r *http.Request) error {
	if u == nil {
		return errors.New("missing user field")
	}
	return nil
}

func (u *Account) Bind(r *http.Request) error {
	if u == nil {
		return errors.New("missing account field")
	}
	return nil
}

func (u *Transaction) Bind(r *http.Request) error {
	if u == nil {
		return errors.New("missing transaction field")
	}
	return nil
}

// TODO Use receiver functions to generalize similar functions?
