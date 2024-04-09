package minipool

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
	"github.com/urfave/cli/v2"
)

func beginReduceBondAmount(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get details
	details, err := rp.Api.Minipool.GetBeginReduceBondDetails()
	if err != nil {
		return err
	}
	if details.Data.BondReductionDisabled {
		fmt.Println("Bond reduction is currently disabled.")
		return nil
	}

	// Check the fee distributor
	if !details.Data.IsFeeDistributorInitialized {
		fmt.Println("Minipools cannot have their bonds reduced until your fee distributor has been initialized.\nPlease run `rocketpool node initialize-fee-distributor` first, then return here to reduce your bonds.")
		return nil
	}

	// TODO POST-ATLAS: Ask the user how much they want the new bond to be; since there's only one option right now there's no point
	fmt.Printf("This will allow you to begin the bond reduction process to reduce your 16 ETH bond for a minipool down to 8 ETH, awarding you 8 ETH in credit and allowing you to create a second minipool for free (plus gas costs).\n\nThere will be a %.0f-hour wait period after you start the process. After this wait period is over, you will have %.0f hours to complete the process. Your `node` container will do this automatically unless you have it disabled, in which case you must manually run `rocketpool minipool reduce-bond`.\n\n%sNOTE: If you don't run it during this window, your request will time out and you will have to start over.%s\n\n", (time.Duration(details.Data.BondReductionWindowStart) * time.Second).Hours(), (time.Duration(details.Data.BondReductionWindowLength) * time.Second).Hours(), terminal.ColorYellow, terminal.ColorReset)
	newBondAmount := eth.EthToWei(8)

	// Prompt for confirmation
	if !(c.Bool("yes") || utils.Confirm("Do you understand how the bond reduction process will work?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Get reduceable minipools
	reduceableMinipools := []api.MinipoolBeginReduceBondDetails{}
	nonReducableMinipoolMessages := []string{}
	for _, minipool := range details.Data.Details {
		if minipool.CanReduce {
			reduceableMinipools = append(reduceableMinipools, minipool)
		} else {
			if minipool.NodeDepositTooLow {
				// Ignore minipools that have already been reduced as much as possible
				continue
			}

			var reason string
			if minipool.AlreadyCancelled {
				reason = "previous bond reduction was scrubbed by the Oracle DAO; no longer eligible"
			} else if minipool.MinipoolVersionTooLow {
				reason = "minipool delegate version too low; upgrade the delegate first"
			} else if minipool.AlreadyInWindow {
				reason = "bond reduction already in progress"
			} else if minipool.BalanceTooLow {
				reason = "minipool has less than 32 ETH on the Beacon Chain"
			} else if minipool.InvalidBeaconState {
				reason = "must be pending or active on the Beacon Chain"
			}
			nonReducableMinipoolMessages = append(nonReducableMinipoolMessages, fmt.Sprintf("%s (%s)", minipool.Address.Hex(), reason))
		}
	}

	// Print ineligible ones
	if len(nonReducableMinipoolMessages) > 0 {
		fmt.Printf("%sNOTE: The following minipools are not eligible for bond reduction:\n", terminal.ColorYellow)
		for _, msg := range nonReducableMinipoolMessages {
			fmt.Println(msg)
		}
		fmt.Printf("%s\n\n", terminal.ColorReset)
	}

	// Check for reduceable minipools
	if len(reduceableMinipools) == 0 {
		fmt.Println("No minipools can have their bond reduced at this time.")
		return nil
	}

	// Get selected minipools
	options := make([]utils.SelectionOption[api.MinipoolBeginReduceBondDetails], len(reduceableMinipools))
	for i, mp := range reduceableMinipools {
		option := &options[i]
		option.Element = &reduceableMinipools[i]
		option.ID = fmt.Sprint(mp.Address)
		option.Display = fmt.Sprintf("%s (Current bond: %d ETH, commission: %.2f%%)", mp.Address.Hex(), int(eth.WeiToEth(mp.NodeDepositBalance)), eth.WeiToEth(mp.NodeFee)*100)
	}
	selectedMinipools, err := utils.GetMultiselectIndices(c, minipoolsFlag, options, "Please select a minipool to begin the ETH bond reduction for:")
	if err != nil {
		return fmt.Errorf("error determining minipool selection: %w", err)
	}

	// Build the TXs
	addresses := make([]common.Address, len(selectedMinipools))
	for i, lot := range selectedMinipools {
		addresses[i] = lot.Address
	}
	response, err := rp.Api.Minipool.BeginReduceBond(addresses, newBondAmount)
	if err != nil {
		return fmt.Errorf("error during TX generation: %w", err)
	}

	// Validation
	totalMatchRequest := big.NewInt(0)
	txs := make([]*eth.TransactionInfo, len(selectedMinipools))
	for i, minipool := range selectedMinipools {
		txInfo := response.Data.TxInfos[i]
		txs[i] = txInfo
		totalMatchRequest.Add(totalMatchRequest, minipool.MatchRequest)
	}

	// Make sure there's enough collateral to cover all of the pending bond reductions
	collateralResponse, err := rp.Api.Node.CheckCollateral()
	if err != nil {
		return fmt.Errorf("error checking the node's total collateral: %w", err)
	}
	totalMatchAvailable := big.NewInt(0).Sub(collateralResponse.Data.EthMatchedLimit, collateralResponse.Data.EthMatched)
	totalMatchAvailable.Sub(totalMatchAvailable, collateralResponse.Data.PendingMatchAmount)
	if totalMatchAvailable.Cmp(totalMatchRequest) < 0 {
		fmt.Printf("You do not have enough RPL staked to support all of the selected bond reductions.\nYou can borrow %.6f more ETH, but are requesting %.6f ETH with these bond reductions.\nIn total, they would bring you below the minimum RPL staking requirement (including the RPL required for any pending bond reductions you've already started).\nYou will have to stake more RPL first.\n", eth.WeiToEth(totalMatchAvailable), eth.WeiToEth(totalMatchRequest))
		return nil
	}

	// Run the TXs
	validated, err := tx.HandleTxBatch(c, rp, txs,
		fmt.Sprintf("Are you sure you want to begin bond reduction for %d minipools from 16 ETH to 8 ETH?", len(selectedMinipools)),
		func(i int) string {
			return fmt.Sprintf("begin-bond-reduce for minipool %s", selectedMinipools[i].Address.Hex())
		},
		"Beginning bond reduction for minipools...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully started bond reduction for all selected minipools.")
	return nil
}
