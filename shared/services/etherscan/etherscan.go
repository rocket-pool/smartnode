package etherscan

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

const gasOracleUrl string = "https://api.etherscan.io/api?module=gastracker&action=gasoracle"

// Standard response
type gasOracleResponse struct {
    Status uinteger             `json:"status"`
    Message string              `json:"message"`
    Result struct {
        SafeGasPrice uinteger       `json:"SafeGasPrice"`
        ProposeGasPrice uinteger    `json:"ProposeGasPrice"`
        FastGasPrice uinteger       `json:"FastGasPrice"`
    }                           `json:"result"`
}

type GasFeeSuggestion struct {
    SlowWei *big.Int
    StandardWei *big.Int
    FastWei *big.Int
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
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return GasFeeSuggestion{}, err
    }

    // Deserialize response
    var gOResponse gasOracleResponse
    if err := json.Unmarshal(body, &gOResponse); err != nil {
        return GasFeeSuggestion{}, fmt.Errorf("Could not decode Etherscan gas oracle response: %w", err)
    }
    if gOResponse.Status != 1 {
        return GasFeeSuggestion{}, fmt.Errorf("Error retrieving Etherscan gas oracle response: %s", gOResponse.Message)
    }

    suggestion := GasFeeSuggestion{
        SlowWei: eth.GweiToWei(float64(gOResponse.Result.SafeGasPrice)),
        StandardWei: eth.GweiToWei(float64(gOResponse.Result.ProposeGasPrice)),
        FastWei: eth.GweiToWei(float64(gOResponse.Result.FastGasPrice)),
    }

    // Return
    return suggestion, nil

}


// Unsigned integer type
type uinteger uint64
func (i uinteger) MarshalJSON() ([]byte, error) {
    return json.Marshal(strconv.Itoa(int(i)))
}

func (i *uinteger) UnmarshalJSON(data []byte) error {

    // Unmarshal string
    var dataStr string
    if err := json.Unmarshal(data, &dataStr); err != nil {
        return err
    }

    // Parse integer value
    value, err := strconv.ParseUint(dataStr, 10, 64)
    if err != nil {
        return err
    }

    // Set value and return
    *i = uinteger(value)
    return nil
}
