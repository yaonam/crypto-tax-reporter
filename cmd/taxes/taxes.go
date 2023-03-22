package taxes

import (
	"crypto-tax-reporter/cmd/models"

	"gorm.io/gorm"
)

func CalculatePNL(db *gorm.DB, userID uint) uint {
	var pnl uint
	// Get sell txs
	var sellTxs []models.Transaction
	db.Where("id == ?", userID).Preload("tax_lots").Find(&sellTxs)
	// Calc PNL by (sell spot price - cost basis) * quantity
	// for _, sellTx := range sellTxs {
	// TODO Add TaxLotSale model, in-between sellTx and TaxLot
	// }
	return pnl
}

func GetTaxLotsFromTxs(db *gorm.DB, accountID uint, txList []models.Transaction) []models.TaxLot {
	var taxLotList []models.TaxLot

	for _, tx := range txList {
		if tx.Type == "buy" {
			var taxLot models.TaxLot
			taxLot.Timestamp = tx.Timestamp
			taxLot.AccountID = accountID
			taxLot.TransactionID = tx.ID
			taxLot.Asset = tx.Asset
			taxLot.Quantity = tx.Quantity
			taxLot.Currency = tx.Currency
			taxLot.CostBasis = tx.Total

			taxLotList = append(taxLotList, taxLot)
			db.Model(&tx).Association("TaxLots").Append(&taxLot)
		} else if tx.Type == "sell" {
			var associatedTaxLots []models.TaxLot
			// Sell tax lots until meet full sell sellQuantity
			for sellQuantity := tx.Quantity; sellQuantity > 0; {
				var taxLot models.TaxLot
				// TODO Make sure taxlot is before sell tx
				db.Where("asset == ? AND is_sold == ?", tx.Asset, false).Order("Timestamp").First(&taxLot)
				taxLotQuantityRemaining := taxLot.Quantity - taxLot.QuantityRealized
				if sellQuantity >= taxLotQuantityRemaining {
					taxLot.QuantityRealized = taxLot.Quantity
					taxLot.IsSold = true
					db.Save(&taxLot)
					sellQuantity -= taxLotQuantityRemaining
				} else {
					taxLot.QuantityRealized += sellQuantity
					db.Save(&taxLot)
					sellQuantity = 0
				}
				associatedTaxLots = append(associatedTaxLots, taxLot)
			}
			db.Model(&tx).Association("TaxLots").Append(&associatedTaxLots)
		}
	}

	return taxLotList
}
