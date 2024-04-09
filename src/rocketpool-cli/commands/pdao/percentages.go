package pdao

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/urfave/cli/v2"
)

func getRewardsPercentages(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get all PDAO settings
	response, err := rp.Api.PDao.RewardsPercentages()
	if err != nil {
		return err
	}

	// Print the settings
	fmt.Printf("Node Operators: %.2f%% (%s)\n", eth.WeiToEth(response.Data.Node)*100, response.Data.Node.String())
	fmt.Printf("Oracle DAO:     %.2f%% (%s)\n", eth.WeiToEth(response.Data.OracleDao)*100, response.Data.OracleDao.String())
	fmt.Printf("Protocol DAO:   %.2f%% (%s)\n", eth.WeiToEth(response.Data.ProtocolDao)*100, response.Data.ProtocolDao.String())
	return nil
}
