package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/urfave/cli/v2"
)

func setSignallingAddress(c *cli.Context, signallingAddress common.Address, signature string) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c)
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.PDao.SetSignallingAddress(signallingAddress, signature)
	if err != nil {
		return fmt.Errorf("Error setting the signalling address: %w", err)
	}

	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to set your signalling address?",
		"setting signalling address",
		"Setting signalling address...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	return nil

}

func clearSignallingAddress(c *cli.Context) error {

	// Get RP client
	rp, err := client.NewClientFromCtx(c)
	if err != nil {
		return err
	}

	// Build the TX
	response, err := rp.Api.PDao.ClearSignallingAddress()
	if err != nil {
		return fmt.Errorf("Error clearing the signalling address: %w", err)
	}

	validated, err := tx.HandleTx(c, rp, response.Data.TxInfo,
		"Are you sure you want to clear the current signalling address?",
		"clearing signalling address",
		"Clearing signalling address...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("The node's signalling address has been sucessfully cleared.")
	return nil
}
