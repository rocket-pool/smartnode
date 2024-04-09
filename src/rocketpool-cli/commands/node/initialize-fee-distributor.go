package node

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
)

func initializeFeeDistributor(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Check if it's already initialized
	response, err := rp.Api.Node.InitializeFeeDistributor()
	if err != nil {
		return err
	}
	if response.Data.IsInitialized {
		fmt.Println("Your fee distributor contract is already initialized.")
		return nil
	}

	fmt.Println("This will create the \"fee distributor\" contract for your node, which captures priority fees and MEV after the merge for you.\n\nNOTE: you don't need to create the contract in order to be given those rewards - you only need it to *claim* those rewards to your withdrawal address.\nThe rewards can accumulate without initializing the contract.\nTherefore, we recommend you wait until the network's gas cost is low to initialize it.")
	fmt.Println()
	fmt.Printf("Your node's fee distributor contract will be created at address %s.\n\n", response.Data.Distributor.Hex())

	// Run the TX
	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to initialize your fee distributor contract?",
		"initializing fee distributor",
		"Initializing fee distributor contract...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Your fee distributor was successfully initialized at address %s.\n", response.Data.Distributor.Hex())
	return nil
}
