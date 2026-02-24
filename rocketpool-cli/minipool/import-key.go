package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/rocketpool-cli/wallet"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/cli/migration"
	"github.com/urfave/cli"
)

func importKey(c *cli.Context, minipoolAddress common.Address) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	fmt.Printf("This will allow you to import the externally-created private key for the validator associated with minipool %s so it can be managed by the Smart Node's Validator Client instead of your externally-managed Validator Client.\n\n", minipoolAddress.Hex())

	// Get the mnemonic
	mnemonic := ""
	if c.IsSet("mnemonic") {
		mnemonic = c.String("mnemonic")
	} else {
		mnemonic = wallet.PromptMnemonic()
	}

	success := migration.ImportKey(c, rp, minipoolAddress, mnemonic)
	if !success {
		fmt.Println("Importing the key failed.\nYou can try again later by using `rocketpool minipool import-key`.")
	}

	return nil
}
