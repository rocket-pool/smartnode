package pdao

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

const (
	colorBlue  string = "\033[36m"
	colorReset string = "\033[0m"
)

func getVotePower(c *cli.Context) error {
	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check for Houston
	houston, err := rp.IsHoustonDeployed()
	if err != nil {
		return fmt.Errorf("error checking if Houston has been deployed: %w", err)
	}
	if !houston.IsHoustonDeployed {
		fmt.Println("This command cannot be used until Houston has been deployed.")
		return nil
	}

	// Get node's voting power at the latest block
	response, err := rp.GetVotingPower()
	if err != nil {
		return err
	}

	status, err := rp.NodeStatus()
	if err != nil {
		return err
	}

	// Print Results
	fmt.Println("== Node Voting Power ==")
	if status.IsVotingInitialized {
		fmt.Println("The node has been initialized for onchain voting.")

	} else {
		fmt.Println("The node has NOT been initialized for onchain voting. You need to run `rocketpool pdao initialize-voting` to participate in onchain votes.")
	}
	if status.OnchainVotingDelegate == status.AccountAddress {
		fmt.Println("The node doesn't have a delegate, which means it can vote directly on onchain proposals.")
	} else {
		fmt.Printf("The node has a voting delegate of %s%s%s which can represent it when voting on Rocket Pool onchain governance proposals.\n", colorBlue, status.OnchainVotingDelegateFormatted, colorReset)
	}

	fmt.Printf("Your current voting power: %.10f\n", eth.WeiToEth(response.VotingPower))
	return nil
}
