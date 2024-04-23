package pdao

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
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

	// Print Results
	fmt.Println("== Node Voting Power ==")
	fmt.Printf("Your current voting power: %d\n", response.VotingPower)
	return nil
}
