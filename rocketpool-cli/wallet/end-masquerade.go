package wallet

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
	"github.com/urfave/cli"
)

func endMasquerade(c *cli.Context) error {
	// Get RP client
	rp := rocketpool.NewClientFromCtx(c)
	defer rp.Close()

	// Get wallet status
	status, err := rp.WalletStatus()
	if err != nil {
		return err
	}

	// Return if not masquerading
	if !status.IsMasquerading {
		fmt.Println("The node is not currently masquerading.")
		return nil
	}

	walletUninitialized := common.Address{}
	if status.NodeAddress == walletUninitialized {
		fmt.Printf("The node wallet is uninitialized. You will no longer be masquerading as %s%s%s.\n\n", colorBlue, status.AccountAddress.Hex(), colorReset)

	} else {
		fmt.Printf("The node wallet is %s%s%s. You will no longer be masquerading as %s%s%s.\n\n", colorBlue, status.NodeAddress, colorReset, colorBlue, status.AccountAddress.Hex(), colorReset)
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to end masquerade mode?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Call Api
	_, err = rp.EndMasquerade()
	if err != nil {
		return fmt.Errorf("error ending masquerade: %w", err)
	}

	fmt.Println("Successfully ended masquerade mode.")

	// Return
	return nil
}
