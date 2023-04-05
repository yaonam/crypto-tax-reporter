package coinbase

import (
	"encoding/csv"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"crypto-tax-reporter/cmd/models"
	"crypto-tax-reporter/cmd/taxes"
)

func Import(db *gorm.DB, accountID uint, filePath string) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// remember to close the file at the end of the program
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	// convert records to array of structs
	txList := parseTxList(db, accountID, data)
	log.Printf("Parsed %v transactions", len(txList))

	// save the array to db
	// TODO: Query to find existing rows, remove from txList, then upload in 2nd query
	var newTxList []models.Transaction
	for _, tx := range txList {
		// result := db.Where("timestamp == ? AND from == ? AND to == ? AND quantity == ?", tx.Timestamp, tx.From, tx.To, tx.Quantity).FirstOrCreate(&tx)
		result := db.Where(&models.Transaction{
			Timestamp: tx.Timestamp,
			From:      tx.From,
			To:        tx.To,
			Quantity:  tx.Quantity,
		}).FirstOrCreate(&tx, tx)
		// result := db.Where(models.Transaction{Timestamp: tx.Timestamp}).FirstOrCreate(&tx)
		if result.RowsAffected == 1 {
			newTxList = append(newTxList, tx)
		}
	}
	log.Printf("Found %v new transactions", len(newTxList))

	// Create tax lots based on txList, mb only use new ones?
	taxes.GetTaxLotsFromTxs(db, accountID, newTxList)
	// Save tax lots to db
	// for _, taxLot := range taxLots {
	// 	db.FirstOrCreate(&taxLot, taxLot)
	// }
}

func parseFloatOrZero(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return 0
}

func parseUintOrZero(s string) uint {
	if f, err := strconv.ParseUint(s, 10, 0); err == nil {
		return uint(f)
	}
	return 0
}

func parseTxList(db *gorm.DB, accountID uint, data [][]string) []models.Transaction {
	var txList []models.Transaction
	for i, line := range data {
		if i > 0 { // skip headers
			// TODO: Handle send
			// Handle based on type
			switch txType := line[1]; txType {
			case "Convert":
				txList = append(txList, handleConvert(db, accountID, line)...)
			case "Learning Reward":
				txList = append(txList, handleReward(db, accountID, line))
			default:
				txList = append(txList, handleBuySell(db, accountID, line))
			}
		}
	}
	return txList
}

func handleBuySell(db *gorm.DB, accountID uint, line []string) models.Transaction {
	// Coinbase columns
	var tx models.Transaction
	tx.Timestamp = line[0]
	switch txType := line[1]; txType {
	case "Buy", "Advanced Trade Buy":
		tx.Type = "buy"
	case "Sell", "Advanced Trade Sell":
		tx.Type = "sell"
	}
	// TODO: Find gas fee when sending to eth wallet
	tx.Asset = models.FindAssetOrCreate(db, line[2])
	tx.Quantity = parseFloatOrZero(line[3])
	tx.Currency = models.FindAssetOrCreate(db, line[4])
	tx.SpotPrice = parseFloatOrZero(line[5])
	tx.Subtotal = parseFloatOrZero(line[6])
	tx.Total = parseFloatOrZero(line[7])
	tx.Fees = parseFloatOrZero(line[8])
	tx.Notes = line[9]

	// Accounts
	tx.From = accountID
	if line[1] == "Send" {
		// Split string
		externalID := strings.Split(line[9], "to ")[1]
		tx.To = models.FindAccountOrCreate(db, accountID, externalID)
	}

	return tx
}

func handleConvert(db *gorm.DB, accountID uint, line []string) []models.Transaction {
	currency := models.FindAssetOrCreate(db, line[4])
	spotPrice := parseFloatOrZero(line[5])
	subtotal := parseFloatOrZero(line[6])
	total := parseFloatOrZero(line[7])
	fees := parseFloatOrZero(line[8])
	notesSplit := strings.Split(line[9], " ")

	// Create sell tx, assign all fees to sell
	var sellTx models.Transaction
	sellTx.Timestamp = line[0]
	sellTx.Type = "sell"
	sellTx.Asset = models.FindAssetOrCreate(db, notesSplit[2])
	sellTx.Quantity = parseFloatOrZero(notesSplit[1])
	sellTx.Currency = currency
	sellTx.SpotPrice = spotPrice
	// Total will be less than subtotal due to fees
	sellTx.Subtotal = total
	sellTx.Total = subtotal
	sellTx.Fees = fees
	sellTx.Notes = line[9]

	// Accounts
	sellTx.From = accountID

	// Create buy tx
	var buyTx models.Transaction
	buyTx.Timestamp = line[0]
	buyTx.Type = "buy"
	buyTx.Asset = models.FindAssetOrCreate(db, notesSplit[5])
	buyTx.Quantity = parseFloatOrZero(notesSplit[4])
	buyTx.Currency = currency
	buyTx.SpotPrice = math.Round(100*subtotal/parseFloatOrZero(notesSplit[4])) / 100
	buyTx.Subtotal = subtotal
	buyTx.Total = subtotal
	buyTx.Fees = 0
	buyTx.Notes = line[9]

	// Accounts
	buyTx.From = accountID

	return []models.Transaction{sellTx, buyTx}
}

func handleReward(db *gorm.DB, accountID uint, line []string) models.Transaction {
	// Create buy tx with 0 cost
	var tx models.Transaction
	tx.Timestamp = line[0]
	tx.Type = "buy"
	tx.Asset = models.FindAssetOrCreate(db, line[2])
	tx.Quantity = parseFloatOrZero(line[3])
	tx.Currency = models.FindAssetOrCreate(db, line[4])
	tx.SpotPrice = parseFloatOrZero(line[5])
	tx.Subtotal = 0
	tx.Total = 0
	tx.Fees = parseFloatOrZero(line[8])
	tx.Notes = line[9]

	// Accounts
	tx.From = accountID

	return tx
}
