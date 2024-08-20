package pdao

import (
	"fmt"
	"strings"

	"github.com/rocket-pool/node-manager-core/utils/input"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	cliutils "github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/urfave/cli/v2"
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
	rp, err := client.NewClientFromCtx(c)
	if err != nil {
		return err
	}

	// Get the TX
	response, err := rp.Api.PDao.InitializeVoting()
	if err != nil {
		return err
	}

	// Verify
	if response.Data.VotingInitialized {
		fmt.Println("Voting has already been initialized for your node.")
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to initialize voting so you can vote on Protocol DAO proposals?",
		"initialize voting",
		"Initializing voting...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully initialized voting. Your node can now vote on Protocol DAO proposals.")
	return nil

}

func initializeVotingWithDelegate(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c)
	if err != nil {
		return err
	}

	// Get the address
	delegateAddressString := c.String("address")
	if delegateAddressString == "" {
		delegateAddressString = utils.Prompt("Please enter the delegate's address:", "^0x[0-9a-fA-F]{40}$", "Invalid member address")
	}
	delegateAddress, err := input.ValidateAddress("delegateAddress", delegateAddressString)
	if err != nil {
		return err
	}

	// Get the TX
	response, err := rp.Api.PDao.InitializeVotingWithDelegate(delegateAddress)
	if err != nil {
		return err
	}

	// Verify
	if response.Data.VotingInitialized {
		fmt.Println("Voting has already been initialized for your node.")
		return nil
	}

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to initialize voting?",
		"initialize voting",
		"Initializing voting...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully initialized voting.")
	return nil
}
