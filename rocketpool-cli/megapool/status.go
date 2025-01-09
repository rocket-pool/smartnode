package megapool

import (
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/urfave/cli"
)

func getStatus(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if Saturn is deployed
	saturnResp, err := rp.IsSaturnDeployed()
	if err != nil {
		return err
	}
	if !saturnResp.IsSaturnDeployed {
		fmt.Println("This command is only available after the Saturn upgrade.")
		return nil
	}

	// Get Megapool status
	status, err := rp.MegapoolStatus()
	if err != nil {
		return err
	}

	fmt.Printf("Node Account Address Formatted %s\n", status.NodeAccountAddressFormatted)
	fmt.Printf("Megapool Address: %s\n", status.Megapool.MegapoolAddress)
	fmt.Printf("Megapool Address Formatted: %s\n", status.Megapool.MegapoolAddressFormatted)
	fmt.Printf("Megapool Deployed: %t\n", status.Megapool.MegapoolDeployed)

	return nil
}
