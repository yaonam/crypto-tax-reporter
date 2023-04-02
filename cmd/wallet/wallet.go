package wallet

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"gorm.io/gorm"
)

const AlchemyApiUrl = "https://eth-mainnet.g.alchemy.com/v2/"

func Import(db *gorm.DB, address string) {
	log.Print("Importing wallet")
	getTransfers()
}

func getTransfers() {
	// Fetch external, internal, and erc20 transfers from Alchemy
	// Set address as To and From
	AlchemyApiKey := os.Getenv("ALCHEMY_API_KEY")

	toParams := map[string]interface{}{
		"fromAddress":  "0x844e94FC29D39840229F6E47290CbE73f187c3b1",
		"category":     []string{"external", "internal", "erc20"},
		"withMetadata": true,
	}
	toData := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "alchemy_getAssetTransfers",
		"params":  []interface{}{toParams},
	}
	toDataJson, _ := json.Marshal(&toData)
	toResp, toErr := http.Post(AlchemyApiUrl+AlchemyApiKey, "application/json", bytes.NewBuffer(toDataJson))
	if toErr != nil {
		log.Fatal("Fetch 'to' transfers from Alchemy failed")
	}
	toBody, toBodyErr := io.ReadAll(toResp.Body)
	if toBodyErr != nil {
		log.Fatal(toBodyErr)
	}
	toResp.Body.Close()
	log.Print(string(toBody))
	// if jsonErr := json.Unmarshal(body, &allApartments); jsonErr != nil {
	// 	log.Fatal(jsonErr)
	// }

	// toJson, toJsonErr := json.Marshal(&toResp)
	// if toJsonErr != nil {
	// 	log.Fatal(toJsonErr)
	// }
	// log.Print(toJson)
}

// func getTxsFromTransfers()
