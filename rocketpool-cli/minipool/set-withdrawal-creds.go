package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/rocketpool-cli/wallet"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)

func printFailureMessage() {
	fmt.Println("Your withdrawal credentials cannot be automatically changed at this time. Import aborted.")
	fmt.Println("You can try again later by using `rocketpool minipool set-withdrawal-creds`.")
}

func setWithdrawalCreds(mnemonic string, minipoolAddress common.Address) error {

	// Get RP client
	rp, err := rocketpool.NewClient().WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	fmt.Printf("This will convert the withdrawal credentials for minipool %s's validator from the old 0x00 (BLS) value to the minipool address. This is meant for solo validator conversion **only**.\n\n", minipoolAddress.Hex())

	// Get the mnemonic
	if mnemonic == "" {
		mnemonic = wallet.PromptMnemonic()
	}

	// Check if the withdrawal creds can be changed
	changeResponse, err := rp.CanChangeWithdrawalCredentials(minipoolAddress, mnemonic)
	if err != nil {
		fmt.Printf("Error checking if withdrawal creds can be migrated: %s\n", err.Error())
		printFailureMessage()
		return nil
	}
	if !changeResponse.CanChange {
		printFailureMessage()
		return nil
	}

	// Change the withdrawal creds
	fmt.Print("Changing withdrawal credentials to the minipool address... ")
	_, err = rp.ChangeWithdrawalCredentials(minipoolAddress, mnemonic)
	if err != nil {
		fmt.Printf("error changing withdrawal credentials: %s\n", err.Error())
		printFailureMessage()
		return nil
	}
	fmt.Println("done!")

	return nil
}
