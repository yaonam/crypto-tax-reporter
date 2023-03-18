package main

import (
	"encoding/csv"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

func openFile() {
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
	txList := parseTxList(1, data)
	log.Printf("Parsed %v transactions", len(txList))

	// save the array to db
	// TODO: Query to find existing rows, remove from txList, then upload in 2nd query
	for _, tx := range txList {
		db.FirstOrCreate(&tx, tx)
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

func parseTxList(userID uint, data [][]string) []Transaction {
	var txList []Transaction
	for i, line := range data {
		if i > 0 { // skip headers
			// TODO: Convert types to lowercase
			// TODO: Create Coinbase struct, parse first?
			// Handle based on type
			switch txType := line[1]; txType {
			case "Convert":
				handleConvert(userID, &txList, line)
			case "Learning Reward":
				handleReward(userID, &txList, line)
			default:
				handleBuySell(userID, &txList, line)
			}
		}
	}
	return txList
}

func handleBuySell(userID uint, txList *[]Transaction, line []string) {
	// Coinbase columns
	var tx Transaction
	tx.Timestamp = line[0]
	tx.Type = line[1]
	tx.Asset = findAssetOrCreate(line[2])
	tx.Quantity = parseFloatOrZero(line[3])
	tx.Currency = findAssetOrCreate(line[4])
	tx.SpotPrice = parseFloatOrZero(line[5])
	tx.Subtotal = parseFloatOrZero(line[6])
	tx.Total = parseFloatOrZero(line[7])
	tx.Fees = parseFloatOrZero(line[8])
	tx.Notes = line[9]

	// Accounts
	tx.From = userID
	if line[1] == "Send" {
		// Split string
		externalID := strings.Split(line[9], "to ")[1]
		tx.To = findAccountOrCreate(userID, externalID)
	}

	*txList = append(*txList, tx)
}

func handleConvert(userID uint, txList *[]Transaction, line []string) {
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
	sellTx.From = userID

	*txList = append(*txList, sellTx)

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
	buyTx.From = userID

	*txList = append(*txList, buyTx)
}

func handleReward(userID uint, txList *[]Transaction, line []string) {
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
	tx.From = userID

	*txList = append(*txList, tx)
}
