package queue

import (
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func getStatus(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	saturnDeployed, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}

	// Get queue status
	status, err := rp.QueueStatus()
	if err != nil {
		return err
	}

	// Print & return
	fmt.Printf("The deposit pool has a balance of %.6f ETH.\n", math.RoundDown(eth.WeiToEth(status.DepositPoolBalance), 6))
	fmt.Printf("There are %d available minipools with a total capacity of %.6f ETH.\n", status.MinipoolQueueLength, math.RoundDown(eth.WeiToEth(status.MinipoolQueueCapacity), 6))

	if saturnDeployed.IsSaturnDeployed {
		var queueDetails api.GetQueueDetailsResponse
		// Get the express ticket count
		queueDetails, err = rp.GetQueueDetails()
		if err != nil {
			return err
		}

		fmt.Println("")
		fmt.Printf("There are %d validator(s) on the express queue.\n", queueDetails.ExpressLength)
		fmt.Printf("There are %d validator(s) on the standard queue.\n", queueDetails.StandardLength)
		fmt.Printf("The express queue rate is %d.\n\n", queueDetails.ExpressRate)
	}

	return nil

}
