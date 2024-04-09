package minipool

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/terminal"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	finalizationThreshold   float64 = 8
	distributeThresholdFlag string  = "threshold"
)

func distributeBalance(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get balance distribution details
	details, err := rp.Api.Minipool.GetDistributeDetails()
	if err != nil {
		return err
	}

	// Sort minipools by status
	eligibleMinipools := []api.MinipoolDistributeDetails{}
	versionTooLowMinipools := []api.MinipoolDistributeDetails{}
	balanceLessThanRefundMinipools := []api.MinipoolDistributeDetails{}
	balanceTooBigMinipools := []api.MinipoolDistributeDetails{}
	finalizationAmount := eth.EthToWei(finalizationThreshold)

	for _, mp := range details.Data.Details {
		if mp.CanDistribute {
			eligibleMinipools = append(eligibleMinipools, mp)
		} else {
			if mp.Version < 3 {
				versionTooLowMinipools = append(versionTooLowMinipools, mp)
			}
			if mp.Balance.Cmp(mp.Refund) == -1 {
				balanceLessThanRefundMinipools = append(balanceLessThanRefundMinipools, mp)
			}
			effectiveBalance := big.NewInt(0).Sub(mp.Balance, mp.Refund)
			if effectiveBalance.Cmp(finalizationAmount) >= 0 {
				balanceTooBigMinipools = append(balanceTooBigMinipools, mp)
			}
		}
	}

	// Print ineligible ones
	if len(versionTooLowMinipools) > 0 {
		fmt.Printf("%sWARNING: The following minipools are using an old delegate and cannot have their rewards safely distributed:\n", terminal.ColorYellow)
		for _, mp := range versionTooLowMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nPlease upgrade the delegate for these minipools using `rocketpool minipool delegate-upgrade` in order to distribute their ETH balances.%s\n\n", terminal.ColorReset)
	}
	if len(balanceLessThanRefundMinipools) > 0 {
		fmt.Printf("%sWARNING: The following minipools have refunds larger than their current balances and cannot be distributed at this time:\n", terminal.ColorYellow)
		for _, mp := range balanceLessThanRefundMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nIf you have recently migrated these minipools from solo validators, please wait until enough rewards have been sent from the Beacon Chain to your minipools to cover your refund amounts.%s\n\n", terminal.ColorReset)
	}
	if len(balanceTooBigMinipools) > 0 {
		fmt.Printf("%sWARNING: The following minipools have over 8 ETH in their balances (after accounting for refunds):\n", terminal.ColorYellow)
		for _, mp := range balanceTooBigMinipools {
			fmt.Printf("\t%s\n", mp.Address)
		}
		fmt.Printf("\nDistributing these minipools will close them, effectively terminating them. If you're sure you want to do this, please use `rocketpool minipool close` on them instead.%s\n\n", terminal.ColorReset)
	}

	if len(eligibleMinipools) == 0 {
		fmt.Println("No minipools are eligible for balance distribution.")
		return nil
	}

	// Filter on the threshold if applicable
	threshold := c.Float64(distributeThresholdFlag)
	if threshold != 0 {
		filteredMps := []api.MinipoolDistributeDetails{}

		for _, mp := range eligibleMinipools {
			var amount float64
			if mp.Status == types.MinipoolStatus_Dissolved {
				amount = math.RoundDown(eth.WeiToEth(mp.Balance), 6)
			} else {
				amount = math.RoundDown(eth.WeiToEth(mp.NodeShareOfDistributableBalance), 6) + math.RoundDown(eth.WeiToEth(mp.Refund), 6)
			}

			if amount > threshold {
				filteredMps = append(filteredMps, mp)
			}
		}

		if len(filteredMps) == 0 {
			fmt.Printf("No minipools have a node operator share larger than the threshold of %.6f ETH.\n", threshold)
			return nil
		}
		eligibleMinipools = filteredMps
	}

	// Sort the minipools by their balance, so the most comes first
	sort.Slice(eligibleMinipools, func(i, j int) bool {
		firstDetails := eligibleMinipools[i]
		secondDetails := eligibleMinipools[j]

		var firstAmount float64
		if firstDetails.Status == types.MinipoolStatus_Dissolved {
			firstAmount = math.RoundDown(eth.WeiToEth(firstDetails.Balance), 6)
		} else {
			firstAmount = math.RoundDown(eth.WeiToEth(firstDetails.NodeShareOfDistributableBalance), 6) + math.RoundDown(eth.WeiToEth(firstDetails.Refund), 6)
		}

		var secondAmount float64
		if secondDetails.Status == types.MinipoolStatus_Dissolved {
			secondAmount = math.RoundDown(eth.WeiToEth(secondDetails.Balance), 6)
		} else {
			secondAmount = math.RoundDown(eth.WeiToEth(secondDetails.NodeShareOfDistributableBalance), 6) + math.RoundDown(eth.WeiToEth(secondDetails.Refund), 6)
		}

		// Sort highest-to-lowest
		return firstAmount > secondAmount
	})

	// Get selected minipools
	options := make([]utils.SelectionOption[api.MinipoolDistributeDetails], len(eligibleMinipools))
	for i, mp := range eligibleMinipools {
		option := &options[i]
		option.Element = &eligibleMinipools[i]
		option.ID = fmt.Sprint(mp.Address)

		if mp.Status == types.MinipoolStatus_Dissolved {
			// Dissolved minipools are a special case
			option.Display = fmt.Sprintf("%s (%.6f ETH available, all of which goes to you)", mp.Address.Hex(), math.RoundDown(eth.WeiToEth(mp.Balance), 6))
		} else {
			option.Display = fmt.Sprintf("%s (%.6f ETH available, %.6f ETH goes to you plus a refund of %.6f ETH)", mp.Address.Hex(), math.RoundDown(eth.WeiToEth(mp.Balance), 6), math.RoundDown(eth.WeiToEth(mp.NodeShareOfDistributableBalance), 6), math.RoundDown(eth.WeiToEth(mp.Refund), 6))
		}
	}
	selectedMinipools, err := utils.GetMultiselectIndices(c, minipoolsFlag, options, "Please select a minipool to distribute:")
	if err != nil {
		return fmt.Errorf("error determining minipool selection: %w", err)
	}

	// Build the TXs
	addresses := make([]common.Address, len(selectedMinipools))
	for i, mp := range selectedMinipools {
		addresses[i] = mp.Address
	}
	response, err := rp.Api.Minipool.Distribute(addresses)
	if err != nil {
		return fmt.Errorf("error during TX generation: %w", err)
	}

	// Validation
	txs := make([]*eth.TransactionInfo, len(selectedMinipools))
	for i := range selectedMinipools {
		txInfo := response.Data.TxInfos[i]
		txs[i] = txInfo
	}

	// Run the TXs
	validated, err := tx.HandleTxBatch(c, rp, txs,
		fmt.Sprintf("Are you sure you want to distribute the ETH balance of %d minipools?", len(selectedMinipools)),
		func(i int) string {
			return fmt.Sprintf("distribution of minipoool %s", selectedMinipools[i].Address.Hex())
		},
		"Distributing balance of minipools...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully distributed the ETH balance of all selected minipools.")
	return nil
}
