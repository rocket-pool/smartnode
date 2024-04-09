package network

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
)

func getRplPrice(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get RPL price
	response, err := rp.Api.Network.RplPrice()
	if err != nil {
		return err
	}

	// Print & return
	fmt.Printf("The current network RPL price is %.6f ETH.\n", math.RoundDown(eth.WeiToEth(response.Data.RplPrice), 6))
	fmt.Printf("Prices last updated at block: %d\n", response.Data.RplPriceBlock)
	return nil
}
