package taxes

import (
	"crypto-tax-reporter/cmd/models"
	"log"
	"math"

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
	db.Model(&account).Where("type = ?", "sell").Preload("TaxLotSales").Preload("TaxLotSales.TaxLot").Association("TxFroms").Find(&sellTxs)
	log.Printf("Length of sell txs: %v", len(sellTxs))
	// Calc PNL by
	for _, sellTx := range sellTxs {
		log.Printf("Calculating tx %v", sellTx.ID)
		for _, taxLotSale := range sellTx.TaxLotSales {
			log.Printf("Calculating tlsale %v", taxLotSale.ID)
			taxLot := taxLotSale.TaxLot
			saleTotal := sellTx.Total * taxLotSale.QuantitySold / sellTx.Quantity
			saleCost := taxLot.CostBasis * taxLotSale.QuantitySold
			salePNL := saleTotal - saleCost
			// Round to nearest hundredth
			salePNL = math.Round(100*salePNL) / 100
			log.Printf("Sale PNL: %v", salePNL)
			pnl += salePNL
		}
	}
	// Round to nearest hundredth
	pnl = math.Round(100*pnl) / 100
	log.Printf("Calculated pnl of %v for account %v", pnl, account.ID)
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
			taxLot.CostBasis = tx.Total / tx.Quantity

			taxLotList = append(taxLotList, taxLot)
			// TODO: Fix taxlots getting assigned tx's way later than their taxlotsale.sellTx
			db.Where("transaction_id == ?", tx.ID).FirstOrCreate(&taxLot)
		} else if tx.Type == "sell" {
			var associatedTaxLots []models.TaxLot
			var taxLotSales []models.TaxLotSale
			// Sell tax lots until meet full sell sellQuantity
			for sellQuantity := tx.Quantity; sellQuantity > 0; {
				var taxLot models.TaxLot
				// TODO Make sure taxlot is before sell tx
				result := db.Where("asset == ? AND is_sold == ?", tx.Asset, false).Order("Timestamp").First(&taxLot)
				if result.Error != nil {
					log.Println("Taxlot not found!")
					break
				}
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
					quantitySold = sellQuantity
					sellQuantity = 0
				}
				associatedTaxLots = append(associatedTaxLots, taxLot)
				taxLotSales = append(taxLotSales, models.TaxLotSale{
					TransactionID: tx.ID,
					TaxLotID:      taxLot.ID,
					QuantitySold:  quantitySold,
				})
			}
			db.Model(&tx).Association("TaxLots").Append(&associatedTaxLots)
			for _, taxLotSale := range taxLotSales {
				db.FirstOrCreate(&taxLotSale, taxLotSale)
			}
		}
	}

	return taxLotList
}
