package migration

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/rocketpool-cli/utils/rocketpool"
)

// Changes a vacant minipool's withdrawal credentials from 0x00 to 0x01
func ChangeWithdrawalCreds(rp *rocketpool.Client, minipoolAddress common.Address, mnemonic string) bool {

	// Check if the withdrawal creds can be changed
	changeResponse, err := rp.CanChangeWithdrawalCredentials(minipoolAddress, mnemonic)
	success := true
	if err != nil {
		fmt.Printf("Error checking if withdrawal creds can be migrated: %s\n", err.Error())
		success = false
	}
	if !changeResponse.CanChange {
		success = false
	}
	if !success {
		return false
	}

	// Change the withdrawal creds
	fmt.Print("Changing withdrawal credentials to the minipool address... ")
	_, err = rp.ChangeWithdrawalCredentials(minipoolAddress, mnemonic)
	if err != nil {
		fmt.Printf("error changing withdrawal credentials: %s\n", err.Error())
		return false
	}
	fmt.Println("done!")
	return true

}
