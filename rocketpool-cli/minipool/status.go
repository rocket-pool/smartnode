package minipool

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/hex"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

const colorReset string = "\033[0m"
const colorRed string = "\033[31m"
const colorYellow string = "\033[33m"

func getStatus(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Get minipool statuses
	status, err := rp.MinipoolStatus()
	if err != nil {
		return err
	}

	// Get minipools by status
	statusMinipools := map[string][]api.MinipoolDetails{}
	refundableMinipools := []api.MinipoolDetails{}
	closeableMinipools := []api.MinipoolDetails{}
	finalisedMinipools := []api.MinipoolDetails{}
	minipoolsPastDissolveNotificationThreshold := []api.MinipoolDetails{}
	for _, minipool := range status.Minipools {

		if !minipool.Finalised {
			// Add to status list
			statusName := minipool.Status.Status.String()
			if _, ok := statusMinipools[statusName]; !ok {
				statusMinipools[statusName] = []api.MinipoolDetails{}
			}
			statusMinipools[statusName] = append(statusMinipools[statusName], minipool)

			// Add to actionable lists
			if minipool.RefundAvailable {
				refundableMinipools = append(refundableMinipools, minipool)
			}
			if minipool.CloseAvailable {
				closeableMinipools = append(closeableMinipools, minipool)
			}
			if minipool.Status.Status == types.Prelaunch && minipool.TimeUntilDissolve.Hours() < minipool.DissolveTimeout.Hours()/2 {
				minipoolsPastDissolveNotificationThreshold = append(minipoolsPastDissolveNotificationThreshold, minipool)
			}
		} else {
			finalisedMinipools = append(finalisedMinipools, minipool)
		}

	}

	// Return if there aren't any minipools
	if len(status.Minipools) == 0 {
		fmt.Println("The node does not have any minipools yet.")
		return nil
	}

	// Return if all minipools are finalized and they are hidden
	if len(status.Minipools) == len(finalisedMinipools) && !c.Bool("include-finalized") {
		fmt.Println("All of this node's minipools have been finalized.\nTo show finalized minipools, re-run this command with the `-f` flag.")
		return nil
	}

	// Print minipool details by status
	for _, statusName := range types.MinipoolStatuses {
		minipools, ok := statusMinipools[statusName]
		if !ok {
			continue
		}

		fmt.Printf("%d %s minipool(s):\n", len(minipools), statusName)
		if statusName == "Withdrawable" {
			fmt.Println("(Withdrawal may not be available until after withdrawal delay)")
		}
		fmt.Println("")

		// Minipools
		for _, minipool := range minipools {
			if !minipool.Finalised || c.Bool("include-finalized") {
				printMinipoolDetails(minipool, status.LatestDelegate)
			}
		}

		fmt.Println("")
	}

	// Handle finalized minipools
	if c.Bool("include-finalized") {
		fmt.Printf("%d finalized minipool(s):\n", len(finalisedMinipools))
		fmt.Println("")

		// Minipools
		for _, minipool := range finalisedMinipools {
			printMinipoolDetails(minipool, status.LatestDelegate)
		}
	} else {
		fmt.Printf("%d finalized minipool(s) (hidden)\n", len(finalisedMinipools))
		fmt.Println("")
	}

	fmt.Println("")

	// Print actionable minipool details
	if len(refundableMinipools) > 0 {
		fmt.Printf("%d minipool(s) have refunds available:\n", len(refundableMinipools))
		for _, minipool := range refundableMinipools {
			fmt.Printf("- %s (%.6f ETH to claim)\n", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Node.RefundBalance), 6))
		}
		fmt.Println("")
	}
	if len(closeableMinipools) > 0 {
		fmt.Printf("%d dissolved minipool(s) can be closed:\n", len(closeableMinipools))
		for _, minipool := range closeableMinipools {
			fmt.Printf("- %s (%.6f ETH to claim)\n", minipool.Address.Hex(), math.RoundDown(eth.WeiToEth(minipool.Balances.ETH), 6))
		}
		fmt.Println("")
	}

	if len(minipoolsPastDissolveNotificationThreshold) > 0 {
		fmt.Printf("%sAttention! %d minipool(s) are close to being dissolved:\n%s", colorRed, len(minipoolsPastDissolveNotificationThreshold), colorReset)
		for _, minipool := range minipoolsPastDissolveNotificationThreshold {
			fmt.Printf("- %s (%s until dissolve)\n", minipool.Address.Hex(), minipool.TimeUntilDissolve)
		}
		fmt.Println("")
	}

	// Return
	return nil

}

func printMinipoolDetails(minipool api.MinipoolDetails, latestDelegate common.Address) {

	fmt.Printf("--------------------\n")
	fmt.Printf("\n")

	// Main details
	fmt.Printf("Address:               %s\n", minipool.Address.Hex())
	if minipool.Penalties == 0 {
		fmt.Println("Penalties:             0")
	} else if minipool.Penalties < 3 {
		fmt.Printf("%sStrikes:               %d%s\n", colorYellow, minipool.Penalties, colorReset)
	} else {
		fmt.Printf("%sInfractions:           %d%s\n", colorRed, minipool.Penalties, colorReset)
	}
	fmt.Printf("Status:                %s\n", minipool.Status.Status.String())
	fmt.Printf("Status updated:        %s\n", minipool.Status.StatusTime.Format(TimeFormat))
	fmt.Printf("Node fee:              %f%%\n", minipool.Node.Fee*100)
	fmt.Printf("Node deposit:          %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.Node.DepositBalance), 6))

	// Queue position
	if minipool.Queue.Position != 0 {
		fmt.Printf("Queue position:        %d\n", minipool.Queue.Position)
	}

	// RP ETH deposit details - prelaunch & staking minipools
	if minipool.Status.Status == types.Prelaunch || minipool.Status.Status == types.Staking {
		totalRewards := big.NewInt(0).Add(minipool.NodeShareOfETHBalance, minipool.Node.RefundBalance)
		if minipool.User.DepositAssigned {
			fmt.Printf("RP ETH assigned:       %s\n", minipool.User.DepositAssignedTime.Format(TimeFormat))
			fmt.Printf("RP deposit:            %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.User.DepositBalance), 6))
		} else {
			fmt.Printf("RP ETH assigned:       no\n")
		}
		fmt.Printf("Minipool Balance (EL): %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.Balances.ETH), 6))
		fmt.Printf("Your portion:          %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.NodeShareOfETHBalance), 6))
		fmt.Printf("Available refund:      %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.Node.RefundBalance), 6))
		fmt.Printf("Total EL rewards:      %.6f ETH\n", math.RoundDown(eth.WeiToEth(totalRewards), 6))
	}

	// Validator details - prelaunch and staking minipools
	if minipool.Status.Status == types.Prelaunch ||
		minipool.Status.Status == types.Staking {
		fmt.Printf("Validator pubkey:      %s\n", hex.AddPrefix(minipool.ValidatorPubkey.Hex()))
		fmt.Printf("Validator index:       %s\n", minipool.Validator.Index)
		if minipool.Validator.Exists {
			if minipool.Validator.Active {
				fmt.Printf("Validator active:      yes\n")
			} else {
				fmt.Printf("Validator active:      no\n")
			}
			fmt.Printf("Beacon balance (CL):   %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.Validator.Balance), 6))
			fmt.Printf("Your portion:          %.6f ETH\n", math.RoundDown(eth.WeiToEth(minipool.Validator.NodeBalance), 6))
		} else {
			fmt.Printf("Validator seen:        no\n")
		}
	}

	// Withdrawal details - withdrawable minipools
	if minipool.Status.Status == types.Withdrawable {
		fmt.Printf("Withdrawal available:  yes\n")
	}

	// Delegate details
	if minipool.UseLatestDelegate {
		fmt.Printf("Use latest delegate:   yes\n")
	} else {
		fmt.Printf("Use latest delegate:   no\n")
	}
	fmt.Printf("Delegate address:      %s\n", cliutils.GetPrettyAddress(minipool.Delegate))
	fmt.Printf("Rollback delegate:     %s\n", cliutils.GetPrettyAddress(minipool.PreviousDelegate))
	fmt.Printf("Effective delegate:    %s\n", cliutils.GetPrettyAddress(minipool.EffectiveDelegate))

	if minipool.EffectiveDelegate != latestDelegate {
		fmt.Printf("%s*Minipool can be upgraded to delegate %s!%s\n", colorYellow, latestDelegate.Hex(), colorReset)
	}

	fmt.Printf("\n")

}
