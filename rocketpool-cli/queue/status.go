package queue

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func getStatus(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get queue status
	status, err := rp.QueueStatus()
	if err != nil {
		return err
	}

	// Print & return
	fmt.Printf("The deposit pool has a balance of %.6f ETH.\n", math.RoundDown(eth.WeiToEth(status.DepositPoolBalance), 6))
	fmt.Printf("There are %d available minipools with a total capacity of %.6f ETH.\n", status.MinipoolQueueLength, math.RoundDown(eth.WeiToEth(status.MinipoolQueueCapacity), 6))
	return nil

}
