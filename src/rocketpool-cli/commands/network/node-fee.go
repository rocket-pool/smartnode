package network

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/urfave/cli/v2"
)

func getNodeFee(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get node fee
	response, err := rp.Api.Network.NodeFee()
	if err != nil {
		return err
	}

	// Print & return
	fmt.Printf("The current network node commission rate is %.2f%%.\n", eth.WeiToEth(response.Data.NodeFee)*100)
	fmt.Printf("Minimum node commission rate: %.2f%%\n", eth.WeiToEth(response.Data.MinNodeFee)*100)
	fmt.Printf("Target node commission rate:  %.2f%%\n", eth.WeiToEth(response.Data.TargetNodeFee)*100)
	fmt.Printf("Maximum node commission rate: %.2f%%\n", eth.WeiToEth(response.Data.MaxNodeFee)*100)
	return nil
}
