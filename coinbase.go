package main

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
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
	txList := parseTxList(data)

	// save the array to db
	db.Create(&txList)
}

func parseFloatOrZero(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return 0
}

func parseTxList(data [][]string) []Transaction {
	var txList []Transaction
	for i, line := range data {
		if i > 0 { // skip headers
			var tx Transaction
			tx.Timestamp = line[0]
			// tx.Type = line[1]
			// tx.Asset = line[2]
			// tx.Quantity = parseFloatOrZero(line[3])
			// tx.Currency = parseFloatOrZero(line[4])
			tx.SpotPrice = parseFloatOrZero(line[5])
			tx.Subtotal = parseFloatOrZero(line[6])
			tx.Total = parseFloatOrZero(line[7])
			tx.Fees = parseFloatOrZero(line[8])
			tx.Notes = line[9]

			txList = append(txList, tx)
		}
	}
	return txList
}
