package faucet

import (
	"fmt"
	"math/big"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
)

func getStatus(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get faucet status
	status, err := rp.Api.Faucet.Status()
	if err != nil {
		return err
	}

	// Print status & return
	fmt.Printf("The faucet has a balance of %.6f legacy RPL.\n", math.RoundDown(eth.WeiToEth(status.Data.Balance), 6))
	if status.Data.WithdrawableAmount.Cmp(big.NewInt(0)) > 0 {
		fmt.Printf("You can withdraw %.6f legacy RPL (requires a %.6f testnet ETH fee)!\n", math.RoundDown(eth.WeiToEth(status.Data.WithdrawableAmount), 6), math.RoundDown(eth.WeiToEth(status.Data.WithdrawalFee), 6))
	} else {
		fmt.Println("You cannot withdraw legacy RPL right now.")
	}
	fmt.Printf("Allowances reset in %d blocks.\n", status.Data.ResetsInBlocks)
	return nil
}
