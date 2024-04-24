package pdao

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/urfave/cli/v2"
)

func getVotePower(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get node's voting power at the latest block
	response, err := rp.Api.PDao.GetVotingPower()
	if err != nil {
		return err
	}

	// Print Results
	fmt.Println("== Node Voting Power ==")
	fmt.Printf("Your current voting power: %.10f\n", eth.WeiToEth(response.Data.VotingPower))
	return nil
}
