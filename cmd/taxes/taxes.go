package taxes

import (
	"crypto-tax-reporter/cmd/models"
	"log"

	"gorm.io/gorm"
)

func CalculateUserPNL(db *gorm.DB, userID uint) float64 {
	var pnl float64
	// Get sell txs
	var userAccounts []models.Account
	db.Where("user = ?", userID).Find(&userAccounts)
	log.Printf("Length of user's account: %v", len(userAccounts))
	for _, account := range userAccounts {
		log.Printf("Calculating pnl for account %v", account.ID)
		pnl += CalculateAccountPNL(db, account)
	}
	return pnl
}

func CalculateAccountPNL(db *gorm.DB, account models.Account) float64 {
	var pnl float64
	// Get sell txs
	var sellTxs []models.Transaction
	db.Model(&account).Where("type = ?", "sell").Association("TxFroms").Find(&sellTxs)
	log.Printf("Length of sell txs: %v", len(sellTxs))
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
				result := db.Where("asset == ? AND is_sold == ?", tx.Asset, false).Order("Timestamp").First(&taxLot)
				if result.Error == nil {
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
				} else {
					log.Println("Taxlot not found!")
					sellQuantity = 0
				}
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
