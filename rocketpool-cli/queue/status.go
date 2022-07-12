package queue

import (
	"fmt"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func getStatus(c *cli.Context) error {

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

	// Get queue status
	status, err := rp.QueueStatus()
	if err != nil {
		return err
	}

	// Print & return
	fmt.Printf("The staking pool has a balance of %.6f ETH.\n", math.RoundDown(eth.WeiToEth(status.DepositPoolBalance), 6))
	fmt.Printf("There are %d available minipools with a total capacity of %.6f ETH.\n", status.MinipoolQueueLength, math.RoundDown(eth.WeiToEth(status.MinipoolQueueCapacity), 6))
	return nil

}
