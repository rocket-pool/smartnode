package node

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)


func nodeStakeRpl(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get node status
    status, err := rp.NodeStatus()
    if err != nil {
        return err
    }

    // Check for fixed-supply RPL balance
    rplBalance := *(status.AccountBalances.RPL)
    if status.AccountBalances.FixedSupplyRPL.Cmp(big.NewInt(0)) > 0 {

        // Confirm swapping RPL
        if (c.Bool("swap") || cliutils.Confirm(fmt.Sprintf("The node has a balance of %.6f old RPL. Would you like to swap it for new RPL before staking?", math.RoundDown(eth.WeiToEth(status.AccountBalances.FixedSupplyRPL), 6)))) {

            // Approve RPL for swapping
            response, err := rp.NodeSwapRplApprove(status.AccountBalances.FixedSupplyRPL)
            if err != nil {
                return err
            }
            hash := response.ApproveTxHash
            fmt.Printf("Approving old RPL for swap...\n")
            cliutils.PrintTransactionHash(hash)
            
            // Swap RPL
            swapResponse, err := rp.NodeSwapRpl(status.AccountBalances.FixedSupplyRPL, hash)
            if err != nil {
                return err
            }
            fmt.Printf("Swapping old RPL for new RPL...\n")
            cliutils.PrintTransactionHash(swapResponse.SwapTxHash)
            if _, err = rp.WaitForTransaction(swapResponse.SwapTxHash); err != nil {
                return err
            }

            // Log
            fmt.Printf("Successfully swapped %.6f old RPL for new RPL.\n", math.RoundDown(eth.WeiToEth(status.AccountBalances.FixedSupplyRPL), 6))
            fmt.Println("")

            // Get new account RPL balance
            rplBalance.Add(status.AccountBalances.RPL, status.AccountBalances.FixedSupplyRPL)

        }

    }

    // Get stake mount
    var amountWei *big.Int
    if c.String("amount") == "min" {

        // Set amount to min per minipool RPL stake
        rplPrice, err := rp.RplPrice()
        if err != nil {
            return err
        }
        amountWei = rplPrice.MinPerMinipoolRplStake

    } else if c.String("amount") == "max" {

        // Set amount to max per minipool RPL stake
        rplPrice, err := rp.RplPrice()
        if err != nil {
            return err
        }
        amountWei = rplPrice.MaxPerMinipoolRplStake

    } else if c.String("amount") == "all" {

        // Set amount to node's entire RPL balance
        amountWei = &rplBalance

    } else if c.String("amount") != "" {

        // Parse amount
        stakeAmount, err := strconv.ParseFloat(c.String("amount"), 64)
        if err != nil {
            return fmt.Errorf("Invalid stake amount '%s': %w", c.String("amount"), err)
        }
        amountWei = eth.EthToWei(stakeAmount)

    } else {

        // Get min/max per minipool RPL stake amounts
        rplPrice, err := rp.RplPrice()
        if err != nil {
            return err
        }
        minAmount := rplPrice.MinPerMinipoolRplStake
        maxAmount := rplPrice.MaxPerMinipoolRplStake

        // Prompt for amount option
        amountOptions := []string{
            fmt.Sprintf("The minimum minipool stake amount (%.6f RPL)?", math.RoundDown(eth.WeiToEth(minAmount), 6)),
            fmt.Sprintf("The maximum effective minipool stake amount (%.6f RPL)?", math.RoundDown(eth.WeiToEth(maxAmount), 6)),
            fmt.Sprintf("Your entire RPL balance (%.6f RPL)?", math.RoundDown(eth.WeiToEth(&rplBalance), 6)),
            "A custom amount",
        }
        selected, _ := cliutils.Select("Please choose an amount of RPL to stake:", amountOptions)
        switch selected {
            case 0: amountWei = minAmount
            case 1: amountWei = maxAmount
            case 2: amountWei = &rplBalance
        }

        // Prompt for custom amount
        if amountWei == nil {
            inputAmount := cliutils.Prompt("Please enter an amount of RPL to stake:", "^\\d+(\\.\\d+)?$", "Invalid amount")
            stakeAmount, err := strconv.ParseFloat(inputAmount, 64)
            if err != nil {
                return fmt.Errorf("Invalid stake amount '%s': %w", inputAmount, err)
            }
            amountWei = eth.EthToWei(stakeAmount)
        }

    }

    // Check RPL can be staked
    canStake, err := rp.CanNodeStakeRpl(amountWei)
    if err != nil {
        return err
    }
    if !canStake.CanStake {
        fmt.Println("Cannot stake RPL:")
        if canStake.InsufficientBalance {
            fmt.Println("The node's RPL balance is insufficient.")
        }
        return nil
    }

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf("Are you sure you want to stake %.6f RPL? Staked RPL can only be withdrawn after a delay.", math.RoundDown(eth.WeiToEth(amountWei), 6)))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Approve RPL for staking
    response, err := rp.NodeStakeRplApprove(amountWei)
    if err != nil {
        return err
    }
    hash := response.ApproveTxHash
    fmt.Printf("Approving RPL for staking...\n")
    cliutils.PrintTransactionHash(hash)

    // Stake RPL
    stakeResponse, err := rp.NodeStakeRpl(amountWei, hash)
    if err != nil {
        return err
    }
    fmt.Printf("Staking RPL...\n")
    cliutils.PrintTransactionHash(stakeResponse.StakeTxHash)
    if _, err = rp.WaitForTransaction(stakeResponse.StakeTxHash); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully staked %.6f RPL.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
    return nil

}

