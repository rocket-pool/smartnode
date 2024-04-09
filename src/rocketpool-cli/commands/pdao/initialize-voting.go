package pdao

import (
	"fmt"

	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/urfave/cli/v2"
)

func initializeVoting(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
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
