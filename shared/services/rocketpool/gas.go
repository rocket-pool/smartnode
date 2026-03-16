package rocketpool

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
)

// Print a warning about the gas estimate for operations that have multiple transactions
func (rp *Client) PrintMultiTxWarning() {

	color.YellowPrintln("NOTE: This operation requires multiple transactions.")

}

// Get the gas price from the latest block
func (c *Client) GetGasPriceFromLatestBlock() (api.GasPriceFromLatestBlockResponse, error) {
	responseBytes, err := c.callAPI("service get-gas-price-from-latest-block")
	if err != nil {
		return api.GasPriceFromLatestBlockResponse{}, fmt.Errorf("Could not get gas price from latest block: %w", err)
	}
	var response api.GasPriceFromLatestBlockResponse
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return api.GasPriceFromLatestBlockResponse{}, fmt.Errorf("Could not decode gas price from latest block response: %w", err)
	}
	if response.Error != "" {
		return api.GasPriceFromLatestBlockResponse{}, fmt.Errorf("Could not get gas price from latest block: %s", response.Error)
	}
	return response, nil
}
