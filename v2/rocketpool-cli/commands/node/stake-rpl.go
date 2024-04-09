package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	swapFlag string = "swap"
)

func nodeStakeRpl(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c).WithReady()
	if err != nil {
		return err
	}

	// Get node status
	status, err := rp.Api.Node.Status()
	if err != nil {
		return err
	}

	// If a custom nonce is set, print the multi-transaction warning
	if rp.Context.Nonce.Cmp(common.Big0) > 0 {
		utils.PrintMultiTransactionNonceWarning()
	}

	// Check for fixed-supply RPL balance
	rplBalance := status.Data.NodeBalances.Rpl
	if status.Data.NodeBalances.Fsrpl.Cmp(big.NewInt(0)) > 0 {
		if c.Bool(swapFlag) || utils.Confirm(fmt.Sprintf("The node has a balance of %.6f legacy RPL. Would you like to swap it for new RPL before staking?", math.RoundDown(eth.WeiToEth(status.Data.NodeBalances.Fsrpl), 6))) {
			err = SwapRpl(c, rp, status.Data.NodeBalances.Fsrpl)
			if err != nil {
				return fmt.Errorf("error swapping legacy RPL: %w", err)
			}

			// Get new account RPL balance
			rplBalance.Add(status.Data.NodeBalances.Rpl, status.Data.NodeBalances.Fsrpl)
		}
	}

	// Get the RPL price
	priceResponse, err := rp.Api.Network.RplPrice()
	if err != nil {
		return fmt.Errorf("error getting RPL price: %w", err)
	}

	// Get stake amount
	var amountWei *big.Int
	switch c.String(amountFlag) {
	case "min8":
		amountWei = priceResponse.Data.MinPer8EthMinipoolRplStake
	case "min16":
		amountWei = priceResponse.Data.MinPer16EthMinipoolRplStake
	case "all":
		amountWei = rplBalance
	case "":
		amountWei, err = promptForRplAmount(priceResponse.Data, rplBalance)
		if err != nil {
			return err
		}
	default:
		// Parse amount
		stakeAmount, err := strconv.ParseFloat(c.String(amountFlag), 64)
		if err != nil {
			return fmt.Errorf("invalid stake amount '%s': %w", c.String(amountFlag), err)
		}
		amountWei = eth.EthToWei(stakeAmount)
	}

	// Build the stake TX
	stakeResponse, err := rp.Api.Node.StakeRpl(amountWei)
	if err != nil {
		return err
	}

	// Verify
	if !stakeResponse.Data.CanStake {
		fmt.Printf("Cannot stake %.6f RPL:\n", eth.WeiToEth(amountWei))
		if stakeResponse.Data.InsufficientBalance {
			fmt.Println("Your node wallet does not currently have this much RPL.")
		}
		return nil
	}

	// Handle boosting the allowance
	if stakeResponse.Data.ApproveTxInfo != nil {
		fmt.Println("Before staking RPL, you must first give the staking contract approval to interact with your RPL.")
		fmt.Println("This only needs to be done once for your node.")

		// If a custom nonce is set, print the multi-transaction warning
		if rp.Context.Nonce.Cmp(common.Big0) > 0 {
			utils.PrintMultiTransactionNonceWarning()
		}

		// Run the Approve TX
		validated, err := tx.HandleTx(c, rp, stakeResponse.Data.ApproveTxInfo,
			"Do you want to let the staking contract interact with your RPL?",
			"approving RPL for staking",
			"Approving RPL for staking...",
		)
		if err != nil {
			return err
		}
		if validated {
			fmt.Println("Successfully approved staking access to RPL.")
		}

		// Build the stake TX once approval is done
		stakeResponse, err = rp.Api.Node.StakeRpl(amountWei)
		if err != nil {
			return err
		}
	}

	// Run the stake TX
	validated, err := tx.HandleTx(c, rp, stakeResponse.Data.StakeTxInfo,
		fmt.Sprintf("Are you sure you want to stake %.6f RPL? You will not be able to unstake this RPL until you exit your validators and close your minipools, or reach over 150%% collateral!", math.RoundDown(eth.WeiToEth(amountWei), 6)),
		"staking RPL",
		"Staking RPL...",
	)
	if err != nil {
		return err
	}
	if !validated {
		return nil
	}

	// Log & return
	fmt.Printf("Successfully staked %.6f RPL.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
	return nil
}

// Prompt the user for the amount of RPL to stake
func promptForRplAmount(priceResponse *api.NetworkRplPriceData, rplBalance *big.Int) (*big.Int, error) {
	// Get min/max per minipool RPL stake amounts
	minAmount8 := priceResponse.MinPer8EthMinipoolRplStake
	minAmount16 := priceResponse.MinPer16EthMinipoolRplStake

	// Prompt for amount option
	var amountWei *big.Int
	amountOptions := []string{
		fmt.Sprintf("The minimum minipool stake amount for an 8-ETH minipool (%.6f RPL)?", math.RoundUp(eth.WeiToEth(minAmount8), 6)),
		fmt.Sprintf("The minimum minipool stake amount for a 16-ETH minipool (%.6f RPL)?", math.RoundUp(eth.WeiToEth(minAmount16), 6)),
		fmt.Sprintf("Your entire RPL balance (%.6f RPL)?", math.RoundDown(eth.WeiToEth(rplBalance), 6)),
		"A custom amount",
	}
	selected, _ := utils.Select("Please choose an amount of RPL to stake:", amountOptions)
	switch selected {
	case 0:
		amountWei = minAmount8
	case 1:
		amountWei = minAmount16
	case 2:
		amountWei = rplBalance
	}

	// Prompt for custom amount
	if amountWei == nil {
		inputAmount := utils.Prompt("Please enter an amount of RPL to stake:", "^\\d+(\\.\\d+)?$", "Invalid amount")
		stakeAmount, err := strconv.ParseFloat(inputAmount, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid stake amount '%s': %w", inputAmount, err)
		}
		amountWei = eth.EthToWei(stakeAmount)
	}
	return amountWei, nil
}
