package pdao

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/urfave/cli"
)

func setSignallingAddress(c *cli.Context, signallingAddress common.Address, signature string) error {

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

	// Get the gas estimation and check if signalling address can be set
	resp, err := rp.CanSetSignallingAddress(signallingAddress, signature)
	if err != nil {
		return fmt.Errorf("error calling can-set-signalling-address: %w", err)
	}

	// Return if there is no signer
	if resp.NodeToSigner == signallingAddress {
		return fmt.Errorf("Could not set snapshot address, signer address already in use.")
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(resp.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to set the signalling address?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Set Signalling Address
	response, err := rp.SetSignallingAddress(signallingAddress, signature)
	if err != nil {
		return fmt.Errorf("error calling set-signalling-address: %w", err)
	}

	fmt.Printf("Setting signalling address...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return fmt.Errorf("error setting signalling address: %w", err)
	}

	// Log & Return
	// fmt.Println("The node's snapshot address was successfully set.")
	fmt.Printf("The node's signalling address was successfully set to %s\n", signallingAddress.String())
	return nil
}

func clearSnapshotAddress(c *cli.Context) error {

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

	// Get the gas estimation and check if snapshot address can be set
	resp, err := rp.CanClearSnapshotAddress()
	if err != nil {
		return fmt.Errorf("error calling can-clear-set-signalling-address: %w", err)
	}

	// Return if there is no signer
	if resp.NodeToSigner == (common.Address{}) {
		return fmt.Errorf("Could not clear snapshot address, no signer set.")
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(resp.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to clear the snapshot address?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Clear Snapshot Address
	response, err := rp.ClearSnapshotAddress()
	if err != nil {
		return fmt.Errorf("error calling clear-signalling-address: %w", err)
	}

	fmt.Printf("Clearing snapshot address...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return fmt.Errorf("error clearing snapshot address: %w", err)
	}

	// Log & return
	fmt.Println("The node's snapshot address was sucessfully cleared.")
	return nil

}
