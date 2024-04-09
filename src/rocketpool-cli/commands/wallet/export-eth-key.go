package wallet

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
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
	if !status.Data.WalletStatus.Wallet.IsLoaded {
		fmt.Println("The node wallet is not loaded and ready for usage. Please run `rocketpool wallet status` for more details.")
		return nil
	}
	if status.Data.WalletStatus.Wallet.Type != wallet.WalletType_Local {
		fmt.Println("This command can only be run on local wallets; hardware wallets cannot have their keys exported.")
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
