package pdao

import (
	"fmt"
	"strings"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func initializeVotingPrompt(c *cli.Context) error {
	fmt.Println("Thanks for initializing your voting power!")
	fmt.Println("")
	fmt.Println("You have two options:")
	fmt.Println("")
	fmt.Println("1. Vote directly (delegate vote power to yourself)")
	fmt.Println("   This will allow you to vote on proposals directly,")
	fmt.Println("   allowing you to personally shape the direction of the protocol.")
	fmt.Println("")
	fmt.Println("2. Delegate your vote")
	fmt.Println("   This will delegate your vote power to someone you trust,")
	fmt.Println("   giving them the power to vote on your behalf. You will have the option to override.")
	fmt.Println("")
	fmt.Printf("You can see a list of existing public delegates at %s,\n", "https://delegates.rocketpool.net")
	fmt.Println("however, you can delegate to any node address.")
	fmt.Println("")
	fmt.Printf("Learn more about how this all works via: %s\n", "https://docs.rocketpool.net/guides/houston/participate#participating-in-on-chain-pdao-proposals")
	fmt.Println("")

	inputString := cliutils.Prompt("Please type `direct` or `delegate` to continue:", "^(?i)(direct|delegate)$", "Please type `direct` or `delegate` to continue:")
	switch strings.ToLower(inputString) {
	case "direct":
		return initializeVoting(c)
	case "delegate":
		return initializeVotingWithDelegate(c)
	}
	return nil

}

func initializeVoting(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	resp, err := rp.CanInitializeVoting()
	if err != nil {
		return fmt.Errorf("error calling get-voting-initialized: %w", err)
	}

	if resp.VotingInitialized {
		fmt.Println("Node voting was already initialized")
		return nil
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(resp.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to initialize voting?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Initialize voting
	response, err := rp.InitializeVoting()
	if err != nil {
		return fmt.Errorf("error calling initialize-voting: %w", err)
	}

	fmt.Printf("Initializing voting...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return fmt.Errorf("error initializing voting: %w", err)
	}

	// Log & return
	fmt.Println("Successfully initialized voting.")
	return nil
}
