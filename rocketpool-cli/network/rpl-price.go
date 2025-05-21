package network

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func getRplPrice(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get RPL price
	response, err := rp.RplPrice()
	if err != nil {
		return err
	}

	// Print & return
	fmt.Printf("The current network RPL price is %.6f ETH.\n", math.RoundDown(eth.WeiToEth(response.RplPrice), 6))
	fmt.Printf("Prices last updated at block: %d\n", response.RplPriceBlock)
	return nil

}
