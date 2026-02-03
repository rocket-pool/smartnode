package node

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

// Config
const (
	defaultMaxNodeFeeSlippage = 0.01 // 1% below current network fee
)

func nodeDeposit(c *cli.Context) error {

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

	if saturnDeployed.IsSaturnDeployed {
		fmt.Println("Saturn 1 is deployed and this command is deprecated. Use `rocketpool megapool deposit` instead.")
		return nil
	}

	fmt.Println("The minipool queue is closed in anticipation of Saturn 1 (launching Feb 18, 2026), when users will be able to create Megapools. See details here: https://saturn.rocketpool.net/")

	return nil
}
