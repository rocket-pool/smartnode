package faucet

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func getStatus(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c)
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check and assign the EC status
	err = cliutils.CheckExecutionClientStatus(rp)
	if err != nil {
		return err
	}

	// Get faucet status
	status, err := rp.FaucetStatus()
	if err != nil {
		return err
	}

	// Print status & return
	fmt.Printf("The faucet has a balance of %.6f legacy RPL.\n", math.RoundDown(eth.WeiToEth(status.Balance), 6))
	if status.WithdrawableAmount.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("You can withdraw %.6f legacy RPL (requires a %.6f GoETH fee)!\n", math.RoundDown(eth.WeiToEth(status.WithdrawableAmount), 6), math.RoundDown(eth.WeiToEth(status.WithdrawalFee), 6))
	} else {
		fmt.Println("You cannot withdraw legacy RPL right now.")
	}
	fmt.Printf("Allowances reset in %d blocks.\n", status.ResetsInBlocks)
	return nil

}
