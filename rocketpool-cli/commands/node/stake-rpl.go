package node

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/urfave/cli/v2"

	"github.com/rocket-pool/node-manager-core/utils/math"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/client"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	cliutils "github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils"
	"github.com/rocket-pool/smartnode/v2/rocketpool-cli/utils/tx"
	"github.com/rocket-pool/smartnode/v2/shared/types/api"
)

const (
	swapFlag               string = "swap"
	stakeRPLWarningMessage string = "NOTE: By staking RPL, your node will automatically initialize voting power to itself. If you would like to delegate your on-chain voting power, you should run the command `rocketpool pdao initialize-voting` before staking RPL."
)

func nodeStakeRpl(c *cli.Context) error {
	// Get RP client
	rp, err := client.NewClientFromCtx(c)
	if err != nil {
		return err
	}

	// Get node status
	status, err := rp.Api.Node.Status()
	if err != nil {
		return err
	}

	// If hotfix is live and voting isn't initialized, display a warning
	err = warnIfVotingUninitialized(rp, c, stakeRPLWarningMessage)
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
	var stakePercent float64

	// Amount flag custom percentage input
	if strings.HasSuffix(c.String("amount"), "%") {
		_, err := fmt.Sscanf(c.String("amount"), "%f%%", &stakePercent)
		if err != nil {
			return fmt.Errorf("invalid percentage format '%s': %w", c.String("amount"), err)
		}
		amountWei = rplStakeForLEB8(eth.EthToWei(stakePercent/100), priceResponse.Data.RplPrice)

	} else if c.String("amount") == "all" {
		// Set amount to node's entire RPL balance
		amountWei = rplBalance

	} else if c.String("amount") != "" {
		// Parse amount
		stakeAmount, err := strconv.ParseFloat(c.String("amount"), 64)
		if err != nil {
			return fmt.Errorf("Invalid stake amount '%s': %w", c.String("amount"), err)
		}
		amountWei = eth.EthToWei(stakeAmount)

	} else {
		amountWei, err = promptForRplAmount(priceResponse.Data, rplBalance, stakePercent)
		if err != nil {
			return err
		}
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
		fmt.Sprintf("Are you sure you want to stake %.6f RPL? You will not be able to unstake this RPL until you exit your validators and close your minipools, or reach %.6f staked RPL (%.0f%% of bonded eth)!",
			math.RoundDown(eth.WeiToEth(amountWei), 6),
			math.RoundDown(eth.WeiToEth(status.Data.MaximumRplStake), 6),
			eth.WeiToEth(status.Data.MaximumStakeFraction)*100),
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
func promptForRplAmount(priceResponse *api.NetworkRplPriceData, rplBalance *big.Int, stakePercent float64) (*big.Int, error) {
	// Get the RPL stake amounts for 5,10,15% borrowed ETH per LEB8
	fivePercentBorrowedPerMinipool := new(big.Int)
	fivePercentBorrowedPerMinipool.SetString("50000000000000000", 10)
	fivePercentBorrowedRplStake := rplStakeForLEB8(fivePercentBorrowedPerMinipool, priceResponse.RplPrice)
	tenPercentBorrowedRplStake := new(big.Int).Mul(fivePercentBorrowedRplStake, big.NewInt(2))
	fifteenPercentBorrowedRplStake := new(big.Int).Mul(fivePercentBorrowedRplStake, big.NewInt(3))

	// Prompt for amount option
	var amountWei *big.Int
	amountOptions := []string{
		fmt.Sprintf("5%% of borrowed ETH (%.6f RPL) for one minipool?", math.RoundUp(eth.WeiToEth(fivePercentBorrowedRplStake), 6)),
		fmt.Sprintf("10%% of borrowed ETH (%.6f RPL) for one minipool?", math.RoundUp(eth.WeiToEth(tenPercentBorrowedRplStake), 6)),
		fmt.Sprintf("15%% of borrowed ETH (%.6f RPL) for one minipool?", math.RoundUp(eth.WeiToEth(fifteenPercentBorrowedRplStake), 6)),
		fmt.Sprintf("Your entire RPL balance (%.6f RPL)?", math.RoundDown(eth.WeiToEth(rplBalance), 6)),
		"A custom amount",
	}
	selected, _ := utils.Select("Please choose an amount of RPL to stake:", amountOptions)
	switch selected {
	case 0:
		amountWei = fivePercentBorrowedRplStake
	case 1:
		amountWei = tenPercentBorrowedRplStake
	case 2:
		amountWei = fifteenPercentBorrowedRplStake
	case 3:
		amountWei = rplBalance
	}

	// Prompt for custom amount or percentage
	if amountWei == nil {
		inputAmountOrPercent := cliutils.Prompt("Please enter an amount of RPL or percentage of borrowed ETH to stake. (e.g '50' for 50 RPL or '5%' for 5% borrowed ETH as RPL):", "^(0|[1-9]\\d*)(\\.\\d+)?%?$", "Invalid amount")
		if strings.HasSuffix(inputAmountOrPercent, "%") {
			_, err := fmt.Sscanf(inputAmountOrPercent, "%f%%", &stakePercent)
			if err != nil {
				return nil, fmt.Errorf("Invalid percentage format '%s': %w", inputAmountOrPercent, err)
			}
			amountWei = rplStakeForLEB8(eth.EthToWei(stakePercent/100), priceResponse.RplPrice)
			fmt.Println(amountWei)
		} else {
			stakeAmount, err := strconv.ParseFloat(inputAmountOrPercent, 64)
			if err != nil {
				return nil, fmt.Errorf("Invalid stake amount '%s': %w", inputAmountOrPercent, err)
			}
			amountWei = eth.EthToWei(stakeAmount)
		}
	}
	return amountWei, nil

}

func rplStakeForLEB8(borrowedPerMinipool *big.Int, rplPrice *big.Int) *big.Int {
	percentBorrowedRplStake := big.NewInt(0)
	percentBorrowedRplStake.Mul(eth.EthToWei(24), borrowedPerMinipool)
	percentBorrowedRplStake.Div(percentBorrowedRplStake, rplPrice)
	percentBorrowedRplStake.Add(percentBorrowedRplStake, big.NewInt(1))
	amountWei := percentBorrowedRplStake

	return amountWei

}
