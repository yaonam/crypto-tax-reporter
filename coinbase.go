package main

import (
	"encoding/csv"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

func openFile(accountID uint) {
	f, err := os.Open("csv/data.csv")
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
	txList := parseTxList(accountID, data)
	log.Printf("Parsed %v transactions", len(txList))

	// save the array to db
	// TODO: Query to find existing rows, remove from txList, then upload in 2nd query
	var newTxList []Transaction
	for _, tx := range txList {
		result := db.FirstOrCreate(&tx, tx)
		newTxList = append(newTxList, tx)
		if result.RowsAffected == 1 {
		}
	}

	// Create tax lots based on txList, mb only use new ones?
	taxLots := getTaxLotsFromTxs(accountID, newTxList)
	// Save tax lots to db
	for _, taxLot := range taxLots {
		db.FirstOrCreate(&taxLot, taxLot)
	}
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

func parseTxList(accountID uint, data [][]string) []Transaction {
	var txList []Transaction
	for i, line := range data {
		if i > 0 { // skip headers
			// TODO: Convert types to lowercase
			// Handle based on type
			switch txType := line[1]; txType {
			case "Convert":
				txList = append(txList, handleConvert(accountID, line)...)
			case "Learning Reward":
				txList = append(txList, handleReward(accountID, line))
			default:
				txList = append(txList, handleBuySell(accountID, line))
			}
		}
	}
	return txList
}

func handleBuySell(accountID uint, line []string) Transaction {
	// Coinbase columns
	var tx Transaction
	tx.Timestamp = line[0]
	switch txType := line[1]; txType {
	case "Buy", "Advanced Trade Buy":
		tx.Type = "buy"
	case "Sell", "Advanced Trade Sell":
		tx.Type = "sell"
	}
	// TODO: Find gas fee when sending to eth wallet
	tx.Asset = findAssetOrCreate(line[2])
	tx.Quantity = parseFloatOrZero(line[3])
	tx.Currency = findAssetOrCreate(line[4])
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
		tx.To = findAccountOrCreate(accountID, externalID)
	}

	return tx
}

func handleConvert(accountID uint, line []string) []Transaction {
	currency := findAssetOrCreate(line[4])
	spotPrice := parseFloatOrZero(line[5])
	subtotal := parseFloatOrZero(line[6])
	total := parseFloatOrZero(line[7])
	fees := parseFloatOrZero(line[8])
	notesSplit := strings.Split(line[9], " ")

	// Create sell tx, assign all fees to sell
	var sellTx Transaction
	sellTx.Timestamp = line[0]
	sellTx.Type = "sell"
	sellTx.Asset = findAssetOrCreate(notesSplit[2])
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
	var buyTx Transaction
	buyTx.Timestamp = line[0]
	buyTx.Type = "buy"
	buyTx.Asset = findAssetOrCreate(notesSplit[5])
	buyTx.Quantity = parseFloatOrZero(notesSplit[4])
	buyTx.Currency = currency
	buyTx.SpotPrice = math.Round(100*subtotal/parseFloatOrZero(notesSplit[4])) / 100
	buyTx.Subtotal = subtotal
	buyTx.Total = subtotal
	buyTx.Fees = 0
	buyTx.Notes = line[9]

	// Accounts
	buyTx.From = accountID

	return []Transaction{sellTx, buyTx}
}

func handleReward(accountID uint, line []string) Transaction {
	// Create buy tx with 0 cost
	var tx Transaction
	tx.Timestamp = line[0]
	tx.Type = "buy"
	tx.Asset = findAssetOrCreate(line[2])
	tx.Quantity = parseFloatOrZero(line[3])
	tx.Currency = findAssetOrCreate(line[4])
	tx.SpotPrice = parseFloatOrZero(line[5])
	tx.Subtotal = 0
	tx.Total = 0
	tx.Fees = parseFloatOrZero(line[8])
	tx.Notes = line[9]

	// Accounts
	tx.From = accountID

	return tx
}

func getTaxLotsFromTxs(accountID uint, txList []Transaction) []TaxLot {
	var taxLotList []TaxLot

	for _, tx := range txList {
		if tx.Type == "buy" {
			var taxLot TaxLot
			taxLot.Timestamp = tx.Timestamp
			taxLot.AccountID = accountID
			taxLot.TransactionID = tx.ID
			taxLot.Asset = tx.Asset
			taxLot.Quantity = tx.Quantity
			taxLot.Currency = tx.Currency
			taxLot.CostBasis = tx.Total

			taxLotList = append(taxLotList, taxLot)
		}
	}

	return taxLotList
}
