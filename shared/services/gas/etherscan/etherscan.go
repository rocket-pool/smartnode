package etherscan

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/goccy/go-json"
)

const gasOracleUrl string = "https://api.etherscan.io/api?module=gastracker&action=gasoracle"

// Standard response
type gasOracleResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

type GasFeeSuggestion struct {
	SlowGwei     float64
	StandardGwei float64
	FastGwei     float64
}

// Get gas prices
func GetGasPrices() (GasFeeSuggestion, error) {

	// Send request
	response, err := http.Get(gasOracleUrl)
	if err != nil {
		return GasFeeSuggestion{}, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	// Check the response code
	if response.StatusCode != http.StatusOK {
		return GasFeeSuggestion{}, fmt.Errorf("request failed with code %d", response.StatusCode)
	}

	// Get response
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return GasFeeSuggestion{}, err
	}

	// Deserialize response
	var gOResponse gasOracleResponse
	if err := json.Unmarshal(body, &gOResponse); err != nil {
		return GasFeeSuggestion{}, fmt.Errorf("Could not decode Etherscan gas oracle response: %w", err)
	}
	if gOResponse.Status != "1" {
		var errMsg string
		if err := json.Unmarshal(gOResponse.Result, &errMsg); err == nil {
			if errMsg == "Max rate limit reached, please use API Key for higher rate limit" {
				return GasFeeSuggestion{}, fmt.Errorf("Rate limit of 1/5sec applied. Try again in a few seconds.")
			}
			return GasFeeSuggestion{}, fmt.Errorf("Etherscan gas oracle request failed: %s", errMsg)
		}
		return GasFeeSuggestion{}, fmt.Errorf("Could not decode Etherscan gas oracle response: %v", gOResponse)
	}

	// Unmarshal result if response is successful
	var result struct {
		LastBlock       string `json:"lastBlock"`
		SafeGasPrice    string `json:"safeGasPrice"`
		ProposeGasPrice string `json:"proposeGasPrice"`
		FastGasPrice    string `json:"fastGasPrice"`
		SuggestBaseFee  string `json:"suggestBaseFee"`
		GasUsedRatio    string `json:"gasUsedRatio"`
	}
	if err := json.Unmarshal(gOResponse.Result, &result); err != nil {
		return GasFeeSuggestion{}, fmt.Errorf("Could not decode Etherscan gas oracle result: %w", err)
	}

	safeGasPriceFloat, err := strconv.ParseFloat(result.SafeGasPrice, 64)
	if err != nil {
		return GasFeeSuggestion{}, fmt.Errorf("invalid SafeGasPrice: %v", err)
	}

	proposeGasPriceFloat, err := strconv.ParseFloat(result.ProposeGasPrice, 64)
	if err != nil {
		return GasFeeSuggestion{}, fmt.Errorf("invalid ProposeGasPrice: %v", err)
	}

	fastGasPriceFloat, err := strconv.ParseFloat(result.FastGasPrice, 64)
	if err != nil {
		return GasFeeSuggestion{}, fmt.Errorf("invalid FastGasPrice: %v", err)
	}

	suggestion := GasFeeSuggestion{
		SlowGwei:     safeGasPriceFloat,
		StandardGwei: proposeGasPriceFloat,
		FastGwei:     fastGasPriceFloat,
	}

	// Return
	return suggestion, nil

}
