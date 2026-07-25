package network

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/math"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func getRplPrice() error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
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
	fmt.Printf("The current network RPL price is %.6f ETH.\n", math.RoundDown(math.WeiToEth(response.RplPrice), 6))
	fmt.Printf("Prices last updated at block: %d\n", response.RplPriceBlock)
	return nil

}
