package rocketpool

import (
	"github.com/rocket-pool/smartnode/rocketpool-cli/cli/color"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Print a warning about the gas estimate for operations that have multiple transactions
func (c *Client) PrintMultiTxWarning() {

	color.YellowPrintln("NOTE: This operation requires multiple transactions.")

}

// Get the gas price from the latest block
func (c *Client) GetGasPriceFromLatestBlock() (api.GasPriceFromLatestBlockResponse, error) {
	return c.callAPI[api.GasPriceFromLatestBlockResponse]("GET", "/api/service/get-gas-price-from-latest-block", nil, "Could not get gas price from latest block")
}
