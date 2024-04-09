package network

import (
	"fmt"

	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/urfave/cli/v2"
)

func getStats(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get network stats
	response, err := rp.Api.Network.Stats()
	if err != nil {
		return err
	}
	activeMinipools := response.Data.InitializedMinipoolCount +
		response.Data.PrelaunchMinipoolCount +
		response.Data.StakingMinipoolCount +
		response.Data.WithdrawableMinipoolCount +
		response.Data.DissolvedMinipoolCount

	// Print & return
	fmt.Printf("%s========== General Stats ==========%s\n", terminal.ColorGreen, terminal.ColorReset)
	fmt.Printf("Total Value Locked:      %.6f ETH\n", eth.WeiToEth(response.Data.TotalValueLocked))
	fmt.Printf("Staking Pool Balance:    %.6f ETH\n", eth.WeiToEth(response.Data.DepositPoolBalance))
	fmt.Printf("Minipool Queue Demand:   %.6f ETH\n", eth.WeiToEth(response.Data.MinipoolCapacity))
	fmt.Printf("Staking Pool ETH Used:   %.2f%%\n\n", eth.WeiToEth(response.Data.StakerUtilization)*100)

	fmt.Printf("%s============== Nodes ==============%s\n", terminal.ColorGreen, terminal.ColorReset)
	fmt.Printf("Current Commission Rate: %.2f%%\n", eth.WeiToEth(response.Data.NodeFee)*100)
	fmt.Printf("Node Count:              %d\n", response.Data.NodeCount)
	fmt.Printf("Active Minipools:        %d\n", activeMinipools)
	fmt.Printf("    Initialized:         %d\n", response.Data.InitializedMinipoolCount)
	fmt.Printf("    Prelaunch:           %d\n", response.Data.PrelaunchMinipoolCount)
	fmt.Printf("    Staking:             %d\n", response.Data.StakingMinipoolCount)
	fmt.Printf("    Withdrawable:        %d\n", response.Data.WithdrawableMinipoolCount)
	fmt.Printf("    Dissolved:           %d\n", response.Data.DissolvedMinipoolCount)
	fmt.Printf("Finalized Minipools:     %d\n\n", response.Data.FinalizedMinipoolCount)

	fmt.Printf("%s========== Smoothing Pool =========%s\n", terminal.ColorGreen, terminal.ColorReset)
	fmt.Printf("Contract Address:        %s%s%s\n", terminal.ColorBlue, response.Data.SmoothingPoolAddress.Hex(), terminal.ColorReset)
	fmt.Printf("Nodes Opted in:          %d\n", response.Data.SmoothingPoolNodes)
	fmt.Printf("Pending Balance:         %.6f\n\n", eth.WeiToEth(response.Data.SmoothingPoolBalance))

	fmt.Printf("%s============== Tokens =============%s\n", terminal.ColorGreen, terminal.ColorReset)
	fmt.Printf("rETH Price (ETH / rETH): %.6f ETH\n", eth.WeiToEth(response.Data.RethPrice))
	fmt.Printf("RPL Price (ETH / RPL):   %.6f ETH\n", eth.WeiToEth(response.Data.RplPrice))
	fmt.Printf("Total RPL staked:        %.6f RPL\n", eth.WeiToEth(response.Data.TotalRplStaked))
	fmt.Printf("Effective RPL staked:    %.6f RPL\n", eth.WeiToEth(response.Data.EffectiveRplStaked))

	return nil
}
