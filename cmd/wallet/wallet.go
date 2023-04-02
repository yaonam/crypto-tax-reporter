package wallet

import (
	"bytes"
	"crypto-tax-reporter/cmd/models"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"gorm.io/gorm"
)

const AlchemyApiUrl = "https://eth-mainnet.g.alchemy.com/v2/"
const DefaultAsset = "USD"

type AlchemyTransfer struct {
	Timestamp string  `json:"metadata.blockTimestamp"`
	From      string  `json:"from"` // Account
	To        string  `json:"to"`   // Account
	Asset     string  `json:"asset"`
	Quantity  float64 `json:"value"`
	Currency  uint    `json:"currency"`
	Notes     string  `json:"hash"`
}

type AlchemyResponse struct {
	Result struct {
		Transfers []AlchemyTransfer `json:"transfers"`
	} `json:"result"`
}

func Import(db *gorm.DB, address string) {
	log.Print("Importing wallet")
	alchTransfers := getTransfers("0x844e94FC29D39840229F6E47290CbE73f187c3b1")
	txs := convertToTx(db, 1, &alchTransfers)
	log.Print(txs)
}

// Calls Alchemy API to get token transfers for account
func getTransfers(walletAddress string) []AlchemyTransfer {
	// Set address as To and From
	AlchemyApiKey := os.Getenv("ALCHEMY_API_KEY")

	toParams := map[string]interface{}{
		"fromAddress":  walletAddress,
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
	toResp.Body.Close()
	if toBodyErr != nil {
		log.Fatal(toBodyErr)
	}

	// Parse data
	var alchResp AlchemyResponse
	rawErr := json.Unmarshal(toBody, &alchResp)
	if rawErr != nil {
		log.Fatal(rawErr)
	}

	return alchResp.Result.Transfers
}

// Convert AlchemyTransfer to models.Transaction
func convertToTx(db *gorm.DB, userID uint, alchTransfers *[]AlchemyTransfer) []models.Transaction {
	// Convert back to Transaction type
	txs := make([]models.Transaction, len(*alchTransfers))
	for i, tf := range *alchTransfers {
		tx := models.Transaction{
			Timestamp: tf.Timestamp,
			Type:      "send",
			From:      models.FindAccountOrCreate(db, userID, tf.From),
			To:        models.FindAccountOrCreate(db, userID, tf.To),
			Asset:     models.FindAssetOrCreate(db, DefaultAsset),
			Quantity:  tf.Quantity,
			Notes:     tf.Notes,
		}
		txs[i] = models.Transaction(tx)
	}
	return txs
}
