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
const DefaultAsset = "usd"

type AlchemyResponse struct {
	Result struct {
		Transfers AlchemyTransfers `json:"transfers"`
	} `json:"result"`
}

// TODO: Add contract address for verification
type ApprovedTokens struct {
	Mapping map[string]string `json:"mapping"`
}

// Imports token transfers from eth address
func Import(db *gorm.DB, userID uint, address string) {
	// Assign account to user
	account := models.Account{User: userID, Type: "wallet", ExternalID: address}
	models.AssignAccountToUser(db, userID, account)

	log.Print("Loading approved tokens")
	approvedTokens := loadApprovedTokens()

	log.Print("Importing wallet")
	alchTransfers := getTransfers(address)

	log.Printf("Filtering %v transfers by approved tokens", len(alchTransfers))
	alchTransfers.filterUnapproved(approvedTokens)

	log.Print("Assigning types")
	alchTransfers.assignTypes(db, userID)

	// log.Print("Fetching spot prices")
	// alchTransfers.getSpotPrices(approvedTokens)

	log.Printf("Matching %v transfers", len(alchTransfers))
	alchTransfers.matchTransfers(address)

	log.Printf("Converting %v transfers to txs", len(alchTransfers))
	convertToTx(db, userID, &alchTransfers, approvedTokens)
	// txs := convertToTx(db, userID, &alchTransfers, approvedTokens)
	// prettyTxs, _ := json.MarshalIndent(txs, "", "  ")
	// log.Print("Transactions: ", string(prettyTxs))

}

// Get mapping of approved tokens to their id from file
func loadApprovedTokens() *ApprovedTokens {
	file, err := os.Open("approvedTokens.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	var approvedTokens ApprovedTokens
	if err := json.Unmarshal(bytes, &approvedTokens); err != nil {
		log.Fatal(err)
	}
	return &approvedTokens
}

// Calls Alchemy API to get token transfers for account
func getTransfers(walletAddress string) AlchemyTransfers {
	// Set address as To and From
	AlchemyApiKey := os.Getenv("ALCHEMY_API_KEY")

	var transfers AlchemyTransfers
	for _, key := range []string{"toAddress", "fromAddress"} {
		params := map[string]interface{}{
			key:            walletAddress,
			"category":     []string{"external", "internal", "erc20"},
			"withMetadata": true,
		}
		data := map[string]interface{}{
			"id":      1,
			"jsonrpc": "2.0",
			"method":  "alchemy_getAssetTransfers",
			"params":  []interface{}{params},
		}
		dataJson, dataJsonErr := json.Marshal(&data)
		if dataJsonErr != nil {
			log.Fatal(dataJsonErr)
		}
		resp, err := http.Post(AlchemyApiUrl+AlchemyApiKey, "application/json", bytes.NewBuffer(dataJson))
		if err != nil {
			log.Fatal(err)
		}
		body, bodyErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if bodyErr != nil {
			log.Fatal(bodyErr)
		}

		// Parse data
		var alchResp AlchemyResponse
		if rawErr := json.Unmarshal(body, &alchResp); rawErr != nil {
			log.Fatal(rawErr)
		}

		transfers = append(transfers, alchResp.Result.Transfers...)
	}

	return transfers
}

// Filter out token transfers that are not in approved list
func (transfers *AlchemyTransfers) filterUnapproved(approvedTokens *ApprovedTokens) {
	var filteredTransfers []AlchemyTransfer
	for _, transfer := range *transfers {
		if approvedTokens.Mapping[string(transfer.Asset)] != "" {
			filteredTransfers = append(filteredTransfers, transfer)
		}
	}
	*transfers = filteredTransfers
}

// Convert AlchemyTransfer to models.Transaction
func convertToTx(db *gorm.DB, userID uint, alchTransfers *AlchemyTransfers, approvedTokens *ApprovedTokens) []models.Transaction {
	// Convert back to Transaction type
	txs := make([]models.Transaction, len(*alchTransfers))
	a := models.Account{User: userID, Type: "wallet"}
	for i, tf := range *alchTransfers {
		tx := models.Transaction{
			Timestamp: tf.Metadata.Timestamp.String(),
			Type:      "send",
			From:      models.FindAccountOrCreate(db, models.Account{User: a.User, Type: a.Type, ExternalID: tf.From}),
			To:        models.FindAccountOrCreate(db, models.Account{User: a.User, Type: a.Type, ExternalID: tf.To}),
			Asset:     models.FindAssetOrCreate(db, string(tf.Asset)),
			Quantity:  tf.Quantity,
			Currency:  models.FindAssetOrCreate(db, DefaultAsset),
			SpotPrice: tf.getSpotPrice(approvedTokens),
			Notes:     tf.Notes,
		}
		txs[i] = tx
	}
	return txs
}
