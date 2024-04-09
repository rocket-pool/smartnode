package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/commands/wallet"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/migration"
	"github.com/urfave/cli/v2"
)

func setWithdrawalCreds(c *cli.Context, minipoolAddress common.Address) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	fmt.Printf("This will convert the withdrawal credentials for minipool %s's validator from the old 0x00 (BLS) value to the minipool address. This is meant for solo validator conversion **only**.\n\n", minipoolAddress.Hex())

	// Get the mnemonic
	mnemonic := ""
	if c.IsSet(utils.MnemonicFlag) {
		mnemonic = c.String(utils.MnemonicFlag)
	} else {
		mnemonic = wallet.PromptMnemonic()
	}

	success := migration.ChangeWithdrawalCreds(rp, minipoolAddress, mnemonic)
	if !success {
		fmt.Println("Your withdrawal credentials cannot be automatically changed at this time. Import aborted.\nYou can try again later by using `rocketpool minipool set-withdrawal-creds`.")
	}

	return nil
}
