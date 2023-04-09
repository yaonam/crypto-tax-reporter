package wallet

import (
	"crypto-tax-reporter/cmd/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

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

func (s *LowercaseString) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	*s = LowercaseString(strings.ToLower(str))
	return nil
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

type CoinGeckoResponse struct {
	MarketData struct {
		CurrentPrice struct {
			USD float64 `json:"usd"`
		} `json:"current_price"`
	} `json:"market_data"`
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
