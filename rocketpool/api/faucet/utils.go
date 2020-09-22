package faucet

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
    "strconv"
    "strings"

    "github.com/ethereum/go-ethereum/common"
)


// Config
const (
    FaucetURL = "https://beta.rocketpool.net/faucet"
    RequestETHAmount = 33
)


// Faucet withdrawal response
type FaucetWithdrawalResponse struct {
    Error string    `json:"error"`
}


// Post a faucet withdrawal request
func postFaucetWithdrawal(token string, address common.Address) (FaucetWithdrawalResponse, error) {

    // Get values by token type
    var amount string
    switch token {
        case "eth":
            amount = strconv.Itoa(RequestETHAmount)
    }

    // Get request params
    params := url.Values{}
    params.Set("type", strings.ToUpper(token))
    params.Set("amount", amount)
    params.Set("account", address.Hex())

    // Send request
    response, err := http.PostForm(FaucetURL, params)
    if err != nil {
        return FaucetWithdrawalResponse{}, fmt.Errorf("Could not send faucet withdrawal request: %w", err)
    }
    defer response.Body.Close()

    // Return on success
    if response.StatusCode == 200 {
        return FaucetWithdrawalResponse{}, nil
    }

    // Get error response body
    responseBody, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return FaucetWithdrawalResponse{}, fmt.Errorf("Could not read faucet withdrawal response: %w", err)
    }

    // Unmarshal error response
    var withdrawalResponse FaucetWithdrawalResponse
    if err := json.Unmarshal(responseBody, &withdrawalResponse); err != nil {
        return FaucetWithdrawalResponse{}, fmt.Errorf("Could not decode faucet withdrawal response: %w", err)
    }

    // Return
    return withdrawalResponse, nil

}

