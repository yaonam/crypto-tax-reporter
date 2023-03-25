package taxes

import (
	"crypto-tax-reporter/cmd/models"
	"log"

	"gorm.io/gorm"
)

func CalculatePNL(db *gorm.DB, userID uint) float64 {
	var pnl float64
	// Get sell txs
	var sellTxs []models.Transaction
	db.Where("from_id == ? AND type == ?", userID, "sell").Preload("TaxLotSales").Preload("TaxLotSales.TaxLot").Find(&sellTxs)
	// Calc PNL by (sell spot price - cost basis) * quantity
	for _, sellTx := range sellTxs {
		log.Printf("Calculating tx %v", sellTx.ID)
		for _, taxLotSale := range sellTx.TaxLotSales {
			log.Printf("Calculating tlsale %v", taxLotSale.ID)
			taxLot := taxLotSale.TaxLot
			pnl += (sellTx.SpotPrice - taxLot.CostBasis) * taxLotSale.QuantitySold
		}
	}
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
			// db.Model(&tx).Association("TaxLots").Append(&taxLot)
			db.Where("transaction_id == ?", tx.ID).FirstOrCreate(&taxLot)
		} else if tx.Type == "sell" {
			var associatedTaxLots []models.TaxLot
			var taxLotSales []models.TaxLotSale
			// Sell tax lots until meet full sell sellQuantity
			for sellQuantity := tx.Quantity; sellQuantity > 0; {
				var taxLot models.TaxLot
				// TODO Make sure taxlot is before sell tx
				db.Where("asset == ? AND is_sold == ?", tx.Asset, false).Order("Timestamp").First(&taxLot)
				taxLotQuantityRemaining := taxLot.Quantity - taxLot.QuantityRealized
				var quantitySold float64
				if sellQuantity >= taxLotQuantityRemaining {
					taxLot.QuantityRealized = taxLot.Quantity
					taxLot.IsSold = true
					db.Save(&taxLot)
					sellQuantity -= taxLotQuantityRemaining
					quantitySold = taxLotQuantityRemaining
				} else {
					taxLot.QuantityRealized += sellQuantity
					db.Save(&taxLot)
					sellQuantity = 0
					quantitySold = sellQuantity
				}
				associatedTaxLots = append(associatedTaxLots, taxLot)
				taxLotSales = append(taxLotSales, models.TaxLotSale{
					TransactionID: tx.ID,
					TaxLotID:      taxLot.ID,
					QuantitySold:  quantitySold,
				})
			}
			db.Model(&tx).Association("TaxLots").Append(&associatedTaxLots)
			// db.FirstOrCreate(&associatedTaxLots)
			for _, taxLotSale := range taxLotSales {
				db.FirstOrCreate(&taxLotSale, taxLotSale)
			}
		}
	}

	return taxLotList
}
