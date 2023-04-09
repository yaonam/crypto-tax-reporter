package wallet

import (
	"bytes"
	"crypto-tax-reporter/cmd/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
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
	Type        string          `json:"type"`
	From        string          `json:"from"`
	To          string          `json:"to"`
	Quantity    float64         `json:"value"`
	Asset       LowercaseString `json:"asset"`
	SpotPrice   float64         `json:"spot_price"`
	Subtotal    float64         `json:"subtotal"`
	Total       float64         `json:"total"`
	Fees        float64         `json:"fees"`
	Category    string          `json:"category"`
	RawContract struct {
		Address string `json:"address"`
	} `json:"rawContract"`
	Notes string `json:"hash"`
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

// Imports token transfers from eth address
func Import(db *gorm.DB, userID uint, address string) {
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

// Assign types based on to/from addresses
func (tfs *AlchemyTransfers) assignTypes(db *gorm.DB, userID uint) {
	// Get user's accounts from db
	var userAccounts []models.Account
	db.Where("user = ?", userID).Find(&userAccounts)

	// Create a map of wallet addresses
	userWallets := make(map[string]bool)
	for _, account := range userAccounts {
		log.Print(account.ExternalID)
		userWallets[account.ExternalID] = true
	}

	// Iterate over transfers and assign types
	for i := range *tfs {
		tf := &(*tfs)[i]
		to, from := userWallets[tf.To], userWallets[tf.From]
		if to && from {
			tf.Type = "send"
		} else if to {
			tf.Type = "buy"
		} else if from {
			tf.Type = "sell"
		}
	}
	after, _ := json.MarshalIndent(*tfs, "", "  ")
	log.Print(string(after))
}

// Convert AlchemyTransfer to models.Transaction
func convertToTx(db *gorm.DB, userID uint, alchTransfers *AlchemyTransfers, approvedTokens *ApprovedTokens) []models.Transaction {
	// Convert back to Transaction type
	txs := make([]models.Transaction, len(*alchTransfers))
	for i, tf := range *alchTransfers {
		tx := models.Transaction{
			Timestamp: tf.Metadata.Timestamp.String(),
			Type:      "send",
			From:      models.FindAccountOrCreate(db, tf.From),
			To:        models.FindAccountOrCreate(db, tf.To),
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
// TODO: Refactor to set *tf's spotprice instead of returning
func (tf *AlchemyTransfer) getSpotPrice(approvedTokens *ApprovedTokens) float64 {
	tokenID := approvedTokens.Mapping[string(tf.Asset)]
	date := tf.Metadata.Timestamp.Format("02-01-2006")
	coinGeckoURL := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%v/history?date=%v&localization=false", tokenID, date)
	var resp *http.Response
	// TODO: Write a generic rate limited request handler
	for {
		var err error
		resp, err = http.Get(coinGeckoURL)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode == 200 || resp.StatusCode != 429 {
			break
		}
		// Try again in 1s if rate limited
		time.Sleep(1)
	}
	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		log.Fatal(bodyErr)
	}
	resp.Body.Close()

	var coinGeckoResponse CoinGeckoResponse
	err := json.Unmarshal(body, &coinGeckoResponse)
	if err != nil {
		log.Fatal(err)
	}

	return coinGeckoResponse.MarketData.CurrentPrice.USD
}

// Fetch spot prices
func (tfs *AlchemyTransfers) getSpotPrices(approvedTokens *ApprovedTokens) {
	for i, tf := range *tfs {
		(*tfs)[i].SpotPrice = tf.getSpotPrice(approvedTokens)
	}
}

func (tfs *AlchemyTransfers) Len() int {
	return len(*tfs)
}

// Sort by timestamp and then tx hash (notes)
func (tfs *AlchemyTransfers) Less(i, j int) bool {
	if (*tfs)[i].Metadata.Timestamp.Before((*tfs)[j].Metadata.Timestamp) {
		return true
	} else if (*tfs)[i].Metadata.Timestamp.Equal((*tfs)[j].Metadata.Timestamp) && (*tfs)[i].Notes < (*tfs)[j].Notes {
		return true
	}
	return false
}
func (tfs *AlchemyTransfers) Swap(i, j int) {
	log.Printf("Swapping tfs %v and %v", i, j)
	(*tfs)[i], (*tfs)[j] = (*tfs)[j], (*tfs)[i]
}

// Sorts and then matches transfers correct missing info.
func (tfs *AlchemyTransfers) matchTransfers(address string) {
	sort.Sort(tfs)
	// Iterate through them, match if same hash and to/from pair
	var matchedTfIDs []int
	for i, tf := range *tfs {
		if len(matchedTfIDs) == 0 || tf.Notes == (*tfs)[matchedTfIDs[0]].Notes {
			// Empty or matched, add to slice
			matchedTfIDs = append(matchedTfIDs, i)
		} else {
			// Doesn't match, process matched and reset
			tfs.handleMatchedTfs(matchedTfIDs, address)

			matchedTfIDs = []int{i}
		}
	}
}

// Fills in missing spot prices
func (tfs *AlchemyTransfers) handleMatchedTfs(matchedTfIDs []int, address string) {
	if len(matchedTfIDs) <= 1 {
		return
	}
	log.Print("Matched: ", matchedTfIDs)
	var missingSpotPriceCount uint
	var missingSpotPriceID uint
	var total int
	for _, i := range matchedTfIDs {
		tf := (*tfs)[i]
		if tf.SpotPrice == 0 {
			missingSpotPriceCount += 1
			missingSpotPriceID = uint(i)
		} else if tf.To == address {
			missingSpotPriceCount = 0
			total += int(tf.Quantity) * int(tf.SpotPrice)
		} else {
			missingSpotPriceCount = 0
			total -= int(tf.Quantity) * int(tf.SpotPrice)
		}
	}
	// Try to fill in missing spot prices using total
	if missingSpotPriceCount == 1 {
		(*tfs)[missingSpotPriceID].SpotPrice = math.Abs(float64(total) / (*tfs)[missingSpotPriceID].Quantity)
	}
}
