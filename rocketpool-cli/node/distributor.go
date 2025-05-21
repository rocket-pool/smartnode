package node

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/bindings/utils/eth"
	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/cli/prompt"
)

func initializeFeeDistributor(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if it's already initialized
	isInitializedResponse, err := rp.IsFeeDistributorInitialized()
	if err != nil {
		return err
	}
	if isInitializedResponse.IsInitialized {
		fmt.Println("Your fee distributor contract is already initialized.")
		return nil
	}

	fmt.Println("This will create the \"fee distributor\" contract for your node, which captures priority fees and MEV after the merge for you.\n\nNOTE: you don't need to create the contract in order to be given those rewards - you only need it to *claim* those rewards to your withdrawal address.\nThe rewards can accumulate without initializing the contract.\nTherefore, we recommend you wait until the network's gas cost is low to initialize it.")

	// Get the gas estimate
	gasResponse, err := rp.GetInitializeFeeDistributorGas()
	if err != nil {
		return err
	}

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(gasResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	fmt.Printf("Your node's fee distributor contract will be created at address %s.\n", gasResponse.Distributor.Hex())

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to initialize your fee distributor contract?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Initialize it
	response, err := rp.InitializeFeeDistributor()
	if err != nil {
		return err
	}

	fmt.Printf("Initializing fee distributor contract...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Printf("Your fee distributor was successfully initialized at address %s.\n", gasResponse.Distributor.Hex())
	return nil

}

func distribute(c *cli.Context) error {

	// Get RP client
	rp, err := rocketpool.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}
	defer rp.Close()

	// Check if it's already initialized
	isInitializedResponse, err := rp.IsFeeDistributorInitialized()
	if err != nil {
		return err
	}
	if !isInitializedResponse.IsInitialized {
		fmt.Println("Your fee distributor has not been initialized yet so you cannot distribute its balance.\nPlease run `rocketpool node initialize-fee-distributor` to create it first.")
		return nil
	}

	// Get the gas estimate
	canDistributeResponse, err := rp.CanDistribute()
	if err != nil {
		return err
	}

	balance := eth.WeiToEth(canDistributeResponse.Balance)
	if balance == 0 {
		fmt.Printf("Your fee distributor does not have any ETH.")
		return nil
	}

	// Print info
	rEthShare := balance - canDistributeResponse.NodeShare
	fmt.Printf("Your fee distributor's balance of %.6f ETH will be distributed as follows:\n", balance)
	fmt.Printf("\tYour withdrawal address will receive %.6f ETH.\n", canDistributeResponse.NodeShare)
	fmt.Printf("\trETH pool stakers will receive %.6f ETH.\n\n", rEthShare)

	// Assign max fees
	err = gas.AssignMaxFeeAndLimit(canDistributeResponse.GasInfo, rp, c.Bool("yes"))
	if err != nil {
		return err
	}

	// Prompt for confirmation
	if !(c.Bool("yes") || prompt.Confirm("Are you sure you want to distribute the ETH from your node's fee distributor?")) {
		fmt.Println("Cancelled.")
		return nil
	}

	// Distribute
	response, err := rp.Distribute()
	if err != nil {
		return err
	}

	fmt.Printf("Distributing rewards...\n")
	cliutils.PrintTransactionHash(rp, response.TxHash)
	if _, err = rp.WaitForTransaction(response.TxHash); err != nil {
		return err
	}

	// Log & return
	fmt.Println("Successfully distributed your fee distributor's balance. Your rewards should arrive in your withdrawal address shortly.")
	return nil

}
