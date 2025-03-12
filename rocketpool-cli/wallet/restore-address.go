package wallet

import (
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/urfave/cli"
)

func restoreAddress(c *cli.Context) error {
	// Get RP client
	rp := rocketpool.NewClientFromCtx(c)
	defer rp.Close()

	// Get wallet status

	// Check if node wallet is loaded

	// if !status.Wallet.IsLoaded {
	// 	fmt.Println("You do not currently have a node wallet loaded, so there is no address to restore. Please see `rocketpool wallet status` for more details.")
	// 	return nil
	// }
	// if status.Wallet.WalletAddress == status.Address.NodeAddress {
	// 	fmt.Println("Your node address is set to your wallet address; you are not currently masquerading.")
	// 	return nil
	// }

	// Compare wallet address with node address

	// Call Api

	// Return
	return nil
}
