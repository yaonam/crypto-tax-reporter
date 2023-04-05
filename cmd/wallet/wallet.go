package wallet

import (
	"bytes"
	"crypto-tax-reporter/cmd/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

const AlchemyApiUrl = "https://eth-mainnet.g.alchemy.com/v2/"
const DefaultAsset = "usd"

type LowercaseString string

type AlchemyTransfer struct {
	Metadata struct {
		Timestamp time.Time `json:"blockTimestamp"`
	} `json:"metadata"`
	From     string          `json:"from"` // Account
	To       string          `json:"to"`   // Account
	Asset    LowercaseString `json:"asset"`
	Quantity float64         `json:"value"`
	Currency uint            `json:"currency"`
	Notes    string          `json:"hash"`
}

type AlchemyTransfers []AlchemyTransfer

type AlchemyResponse struct {
	Result struct {
		Transfers AlchemyTransfers `json:"transfers"`
	} `json:"result"`
}

// TODO: Add contract address for verification
type ApprovedTokens struct {
	Mapping map[string]string `json:"mapping"`
}

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
	err = json.Unmarshal(bytes, &approvedTokens)
	if err != nil {
		log.Fatal(err)
	}
	return &approvedTokens
}

// Imports token transfers from eth address
func Import(db *gorm.DB, userID uint, address string) {
	log.Print("Loading approved tokens")
	approvedTokens := loadApprovedTokens()

	log.Print("Importing wallet")
	alchTransfers := getTransfers("0x844e94FC29D39840229F6E47290CbE73f187c3b1")

	log.Printf("Filtering %v transfers by approved tokens", len(alchTransfers))
	alchTransfers.filterUnapproved(approvedTokens)

	log.Printf("Converting %v transfers to txs", len(alchTransfers))
	txs := convertToTx(db, userID, &alchTransfers, approvedTokens)
	prettyTxs, _ := json.MarshalIndent(txs, "", "  ")
	log.Print("Transactions: ", string(prettyTxs))
}

func (s *LowercaseString) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	*s = LowercaseString(strings.ToLower(str))
	return nil
}

// Calls Alchemy API to get token transfers for account
func getTransfers(walletAddress string) AlchemyTransfers {
	// Set address as To and From
	AlchemyApiKey := os.Getenv("ALCHEMY_API_KEY")

	var transfers AlchemyTransfers
	toParams := map[string]interface{}{
		"toAddress":    walletAddress,
		"category":     []string{"external", "internal", "erc20"},
		"withMetadata": true,
	}
	fromParams := map[string]interface{}{
		"fromAddress":  walletAddress,
		"category":     []string{"external", "internal", "erc20"},
		"withMetadata": true,
	}
	for _, params := range []interface{}{toParams, fromParams} {
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
		rawErr := json.Unmarshal(body, &alchResp)
		if rawErr != nil {
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
	for i, tf := range *alchTransfers {
		tx := models.Transaction{
			Timestamp: tf.Metadata.Timestamp.String(),
			Type:      "send",
			From:      models.FindAccountOrCreate(db, userID, tf.From),
			To:        models.FindAccountOrCreate(db, userID, tf.To),
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

type CoinGeckoResponse struct {
	MarketData struct {
		CurrentPrice struct {
			USD float64 `json:"usd"`
		} `json:"current_price"`
	} `json:"market_data"`
}

// Fetch spot price from coingecko
func (tf *AlchemyTransfer) getSpotPrice(approvedTokens *ApprovedTokens) float64 {
	tokenID := approvedTokens.Mapping[string(tf.Asset)]
	date := tf.Metadata.Timestamp.Format("02-01-2006")
	coinGeckoURL := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%v/history?date=%v&localization=false", tokenID, date)
	resp, err := http.Get(coinGeckoURL)
	if err != nil {
		log.Fatal(err)
	}
	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		log.Fatal(bodyErr)
	}
	resp.Body.Close()

	var coinGeckoResponse CoinGeckoResponse
	err = json.Unmarshal(body, &coinGeckoResponse)
	if err != nil {
		log.Fatal(err)
	}

	return coinGeckoResponse.MarketData.CurrentPrice.USD
}
