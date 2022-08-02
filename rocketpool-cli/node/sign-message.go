package node

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

func signMessage(c *cli.Context, message string) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get & check wallet status
	status, err := rp.WalletStatus()
	if err != nil {
		return err
	}

	if !status.WalletInitialized {
		fmt.Println("The node wallet is not initialized.")
		return nil
	}

	if message == "" {
		message = cliutils.Prompt("Please enter the message you want to sign: (EIP-191 personal_sign)", "^.+$", "Please enter the message you want to sign: (EIP-191 personal_sign)")
	}

	response, err := rp.SignMessage(message)
	if err != nil {
		return err
	}

	fmt.Printf(`{ 
    "address": "%s",
    "msg": "%s",
    "sig": "%s",
    "version": "2"
}`, status.AccountAddress, message, response.SignedData)

	return nil

}
