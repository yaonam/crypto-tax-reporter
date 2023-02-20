package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"gorm.io/gorm"
)

type CoinbaseTxModel struct {
	gorm.Model
	Timestamp         string
	Transaction       string
	Asset             string
	Quantity          float64
	SpotPriceCurrency float64
	SpotPriceAtTx     float64
	Subtotal          float64
	Total             float64
	Fees              float64
	Notes             string
}

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

	// print the array
	fmt.Printf("%+v\n", txList)
}

func parseFloatOrZero(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return 0
}

func parseTxList(data [][]string) []CoinbaseTxModel {
	var txList []CoinbaseTxModel
	for i, line := range data {
		if i > 0 { // skip headers
			var tx CoinbaseTxModel
			tx.Timestamp = line[0]
			tx.Transaction = line[1]
			tx.Asset = line[2]
			tx.Quantity = parseFloatOrZero(line[3])
			tx.SpotPriceCurrency = parseFloatOrZero(line[4])
			tx.SpotPriceAtTx = parseFloatOrZero(line[5])
			tx.Subtotal = parseFloatOrZero(line[6])
			tx.Total = parseFloatOrZero(line[7])
			tx.Fees = parseFloatOrZero(line[8])
			tx.Notes = line[9]

			txList = append(txList, tx)
		}
	}
	return txList
}
