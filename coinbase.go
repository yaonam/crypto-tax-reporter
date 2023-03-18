package main

import (
	"encoding/csv"
	"log"
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

	// convert records to array of structs
	txList := parseTxList(1, data)
	log.Printf("Parsed %v transactions", len(txList))

	// save the array to db
	for _, tx := range txList {
		db.FirstOrCreate(&tx, tx)
	}
	// db.Create(&txList)
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
			// Handle based on type
			// select case...

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

			txList = append(txList, tx)
		}
	}
	return txList
}

// TODO: Figure out if I need to make args pointers
func handleConvert(userID uint, txList []Transaction, line []string) {
	notesSplit := strings.Split(line[9], " ")

	// Create sell sellTx
	var sellTx Transaction
	sellTx.Timestamp = line[0]
	sellTx.Type = line[1]
	sellTx.Asset = findAssetOrCreate(notesSplit[2])
	sellTx.Quantity = parseFloatOrZero(notesSplit[1])
	sellTx.Currency = findAssetOrCreate(line[4])
	sellTx.SpotPrice = parseFloatOrZero(line[5])
	// Total will be less than subtotal due to fees
	sellTx.Subtotal = parseFloatOrZero(line[7])
	sellTx.Total = parseFloatOrZero(line[6])
	sellTx.Fees = parseFloatOrZero(line[8])
	sellTx.Notes = line[9]

	// Accounts
	sellTx.From = userID

	txList = append(txList, sellTx)

	// Create buy tx
	var buyTx Transaction
	buyTx.Timestamp = line[0]
	buyTx.Type = line[1]
	buyTx.Asset = findAssetOrCreate(notesSplit[5])
	buyTx.Quantity = parseFloatOrZero(notesSplit[4])
	buyTx.Currency = findAssetOrCreate(line[4])
	buyTx.SpotPrice = parseFloatOrZero(line[5])
	buyTx.Subtotal = parseFloatOrZero(line[6])
	buyTx.Total = parseFloatOrZero(line[6])
	buyTx.Fees = 0
	buyTx.Notes = line[9]

	// Accounts
	buyTx.From = userID

	txList = append(txList, buyTx)
}

func handleReward() {}
