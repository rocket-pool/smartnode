package pdao

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/urfave/cli/v2"
)

func getVotePower(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get node's voting power at the latest block
	vpResponse, err := rp.Api.PDao.GetVotingPower()
	if err != nil {
		return err
	}

	// Get the node's status
	statusResponse, err := rp.Api.Node.Status()
	if err != nil {
		return err
	}

	// Print Results
	fmt.Println("== Node Voting Power ==")
	if statusResponse.Data.IsVotingInitialized {
		fmt.Println("The node has been initialized for onchain voting.")

	} else {
		fmt.Println("The node has NOT been initialized for onchain voting. You need to run `rocketpool pdao initialize-voting` to participate in onchain votes.")
	}
	if statusResponse.Data.OnchainVotingDelegate == statusResponse.Data.AccountAddress {
		fmt.Println("The node doesn't have a delegate, which means it can vote directly on onchain proposals.")
	} else {
		fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool onchain governance proposals.\n", terminal.ColorBlue, statusResponse.Data.OnchainVotingDelegateFormatted, terminal.ColorReset)
	}

	fmt.Printf("Your current voting power: %.10f\n", eth.WeiToEth(vpResponse.Data.VotingPower))
	return nil
}
