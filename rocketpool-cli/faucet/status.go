package faucet

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

func getStatus(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get faucet status
	status, err := rp.FaucetStatus()
	if err != nil {
		return err
	}

	// Print status & return
	fmt.Printf("The faucet has a balance of %.6f legacy RPL.\n", math.RoundDown(eth.WeiToEth(status.Balance), 6))
	if status.WithdrawableAmount.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("You can withdraw %.6f legacy RPL (requires a %.6f testnet ETH fee)!\n", math.RoundDown(eth.WeiToEth(status.WithdrawableAmount), 6), math.RoundDown(eth.WeiToEth(status.WithdrawalFee), 6))
	} else {
		fmt.Println("You cannot withdraw legacy RPL right now.")
	}
	fmt.Printf("Allowances reset in %d blocks.\n", status.ResetsInBlocks)
	return nil

}
