package queue

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
)

func getStatus(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get queue status
	status, err := rp.Api.Queue.Status()
	if err != nil {
		return err
	}

	// Print & return
	fmt.Printf("The staking pool has a balance of %.6f ETH.\n", math.RoundDown(eth.WeiToEth(status.Data.DepositPoolBalance), 6))
	fmt.Printf("There are %d available minipools with a total capacity of %.6f ETH.\n", status.Data.MinipoolQueueLength, math.RoundDown(eth.WeiToEth(status.Data.MinipoolQueueCapacity), 6))
	return nil
}
