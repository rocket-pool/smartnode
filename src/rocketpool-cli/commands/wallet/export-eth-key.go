package wallet

import (
	"fmt"

	"github.com/rocket-pool/smartnode/rocketpool-cli/client"
	"github.com/urfave/cli/v2"
)

func exportEthKey(c *cli.Context) error {
	// Get RP client
	rp := client.NewClientFromCtx(c)

	// Get & check wallet status
	status, err := rp.Api.Wallet.Status()
	if err != nil {
		return err
	}
	if !status.Data.WalletStatus.HasKeystore {
		fmt.Println("The node wallet is not initialized.")
		return nil
	}

	// Get the wallet in ETH key format
	ethKey, err := rp.Api.Wallet.ExportEthKey()
	if err != nil {
		return err
	}

	// Print wallet & return
	fmt.Println("Wallet in ETH Key Format:")
	fmt.Println(string(ethKey.Data.EthKeyJson))
	return nil
}
