package network

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func getRplPrice(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check and assign the EC status
	err = cliutils.CheckClientStatus(rp)
	if err != nil {
		return err
	}

	// Get RPL price
	response, err := rp.RplPrice()
	if err != nil {
		return err
	}

	// Print & return
	fmt.Printf("The current network RPL price is %.6f ETH.\n", math.RoundDown(eth.WeiToEth(&response.RplPrice), 6))
	fmt.Printf("Prices last updated at block: %d\n", response.RplPriceBlock)
	return nil

}
