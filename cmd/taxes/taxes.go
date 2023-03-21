package taxes

import (
	"crypto-tax-reporter/cmd/models"

	"gorm.io/gorm"
)

func CalculatePNL(db *gorm.DB, userID uint) uint {
	// Get sell txs
	var sellTxs []models.Transaction
	db.Where("id == ?", userID).Preload("tax_lots").Find(&sellTxs)
	// Calc PNL by (sell spot price - cost basis) * quantity
	// for _, sellTx := range sellTxs {}
	return 0
}
