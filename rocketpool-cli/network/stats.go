package network

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/utils/cli/color"
)

func getStats(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	fmt.Println("Getting network stats. This may take a few seconds...")
	fmt.Println()

	// Get network stats
	response, err := rp.NetworkStats()
	if err != nil {
		return err
	}
	activeMinipools := response.InitializedMinipoolCount +
		response.PrelaunchMinipoolCount +
		response.StakingMinipoolCount +
		response.WithdrawableMinipoolCount +
		response.DissolvedMinipoolCount

	// Print & return
	color.GreenPrintln("========== General Stats ==========")
	fmt.Printf("Total Value Locked:          %f ETH\n", response.TotalValueLocked)
	fmt.Printf("Deposit Pool Balance:        %f ETH\n", response.DepositPoolBalance)
	fmt.Printf("Minipool Queue Demand:       %f ETH\n", response.MinipoolCapacity)
	fmt.Printf("Deposit Pool ETH Used:       %f%%\n", response.StakerUtilization*100)
	fmt.Println()

	color.GreenPrintln("============== Nodes ==============")
	fmt.Printf("Current Commission Rate:     %f%%\n", response.NodeFee*100)
	fmt.Printf("Node Count:                  %d\n", response.NodeCount)
	fmt.Printf("Active Minipools:            %d\n", activeMinipools)
	fmt.Printf("    Initialized:             %d\n", response.InitializedMinipoolCount)
	fmt.Printf("    Prelaunch:               %d\n", response.PrelaunchMinipoolCount)
	fmt.Printf("    Staking:                 %d\n", response.StakingMinipoolCount)
	fmt.Printf("    Withdrawable:            %d\n", response.WithdrawableMinipoolCount)
	fmt.Printf("    Dissolved:               %d\n", response.DissolvedMinipoolCount)
	fmt.Printf("Finalized Minipools:         %d\n", response.FinalizedMinipoolCount)
	fmt.Println()

	color.GreenPrintln("=========== Megapools ============")
	fmt.Printf("Megapool contracts deployed: %d\n", response.MegapoolContractCount)
	fmt.Printf("Total megapool validators:  %d\n", response.MegapoolValidatorCount)
	fmt.Printf("     Staking:                %d\n", response.MegapoolValidatorStakingCount)
	fmt.Printf("     In Prestake:            %d\n", response.MegapoolValidatorInPrestakeCount)
	fmt.Printf("     In Queue:               %d\n", response.MegapoolValidatorInQueueCount)
	fmt.Printf("     Exited:                 %d\n", response.MegapoolValidatorExitedCount)
	fmt.Printf("     Locked:                 %d\n", response.MegapoolValidatorLockedCount)
	fmt.Printf("     Exiting:                %d\n", response.MegapoolValidatorExitingCount)
	fmt.Printf("     Dissolved:              %d\n", response.MegapoolValidatorDissolvedCount)
	fmt.Println()

	color.GreenPrintln("========== Smoothing Pool ==========")
	fmt.Printf("Contract Address:            %s\n", color.LightBlue(response.SmoothingPoolAddress.Hex()))
	fmt.Printf("Nodes Opted in:              %d\n", response.SmoothingPoolNodes)
	fmt.Printf("Pending Balance:             %f\n", response.SmoothingPoolBalance)
	fmt.Println()

	color.GreenPrintln("============== Tokens ==============")
	fmt.Printf("rETH Price (ETH / rETH):     %f ETH\n", response.RethPrice)
	fmt.Printf("RPL Price (ETH / RPL):       %f ETH\n", response.RplPrice)
	fmt.Printf("Total RPL staked:            %f RPL\n", response.TotalRplStaked)
	fmt.Printf("Total Megapool RPL staked:   %f RPL\n", response.TotalMegapoolRplStaked)
	fmt.Printf("Total Legacy RPL staked:     %f RPL\n", response.TotalLegacyRplStaked)
	fmt.Println()

	return nil

}
