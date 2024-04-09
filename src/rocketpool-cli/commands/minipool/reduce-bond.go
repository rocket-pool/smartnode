package minipool

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
	"github.com/urfave/cli/v2"
)

func reduceBondAmount(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get details
	details, err := rp.Api.Minipool.GetReduceBondDetails()
	if err != nil {
		return err
	}

	// Check the fee distributor
	if !details.Data.IsFeeDistributorInitialized {
		fmt.Println("Minipools cannot have their bonds reduced until your fee distributor has been initialized.\nPlease run `rocketpool node initialize-fee-distributor` first, then return here to reduce your bonds.")
		return nil
	}

	fmt.Println("NOTE: this function is used to complete the bond reduction process for a minipool. If you haven't started the process already, please run `rocketpool minipool begin-bond-reduction` first.\n")

	// Get reduceable minipools
	reduceableMinipools := []api.MinipoolReduceBondDetails{}
	for _, minipool := range details.Data.Details {
		if minipool.CanReduce {
			reduceableMinipools = append(reduceableMinipools, minipool)
		}
	}

	// Check for reduceable minipools
	if len(reduceableMinipools) == 0 {
		fmt.Println("No minipools are eligible for bond reduction at this time.")
		return nil
	}

	// Workaround for the fee distribution issue
	err = forceFeeDistribution(c, rp)
	if err != nil {
		return err
	}

	// Get selected minipools
	options := make([]utils.SelectionOption[api.MinipoolReduceBondDetails], len(reduceableMinipools))
	for i, mp := range reduceableMinipools {
		option := &options[i]
		option.Element = &reduceableMinipools[i]
		option.ID = fmt.Sprint(mp.Address)
		option.Display = fmt.Sprintf("%s (Current bond: %d ETH, commission: %.2f%%)", mp.Address.Hex(), int(eth.WeiToEth(mp.NodeDepositBalance)), eth.WeiToEth(mp.NodeFee)*100)
	}
	selectedMinipools, err := utils.GetMultiselectIndices(c, minipoolsFlag, options, "Please select a minipool to reduce the ETH bond for:")
	if err != nil {
		return fmt.Errorf("error determining minipool selection: %w", err)
	}

	// Build the TXs
	addresses := make([]common.Address, len(selectedMinipools))
	for i, mp := range selectedMinipools {
		addresses[i] = mp.Address
	}
	response, err := rp.Api.Minipool.ReduceBond(addresses)
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
		fmt.Sprintf("Are you sure you want to reduce the bond for %d minipools from 16 ETH to 8 ETH?", len(selectedMinipools)),
		func(i int) string {
			return fmt.Sprintf("bond reduction for minipool %s", selectedMinipools[i].Address.Hex())
		},
		"Reducing bond for minipools...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully reduced bond for all selected minipools.")
	return nil
}

func forceFeeDistribution(c *cli.Context, rp *client.Client) error {
	// Get the gas estimate
	response, err := rp.Api.Node.Distribute()
	if err != nil {
		return err
	}

	balance := response.Data.Balance
	if balance.Cmp(common.Big0) == 0 {
		fmt.Println("Your fee distributor does not have any ETH and does not need to be distributed.\n")
		return nil
	}
	fmt.Println("NOTE: prior to bond reduction, you must distribute the funds in your fee distributor.\n")

	// Print info
	balanceFloat := eth.WeiToEth(response.Data.Balance)
	nodeShareFloat := eth.WeiToEth(response.Data.NodeShare)
	rEthShare := balanceFloat - nodeShareFloat
	fmt.Printf("Your fee distributor's balance of %.6f ETH will be distributed as follows:\n", balance)
	fmt.Printf("\tYour withdrawal address will receive %.6f ETH.\n", nodeShareFloat)
	fmt.Printf("\trETH pool stakers will receive %.6f ETH.\n\n", rEthShare)

	// Run the TX
	txInfo := response.Data.TxInfo
	validated, err := tx.HandleTx(c, rp, txInfo,
		"Are you sure you want to distribute the ETH from your node's fee distributor?",
		"node rewards distribution",
		"Distributing rewards...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Println("Successfully distributed your fee distributor's balance. Your rewards should arrive in your withdrawal address shortly.")
	return nil
}
