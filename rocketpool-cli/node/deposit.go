package node

import (
    "fmt"
    "strconv"

    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
    "github.com/rocket-pool/smartnode/shared/utils/math"
)


func nodeDeposit(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get deposit amount
    var amount float64
    if c.String("amount") != "" {

        // Parse amount
        depositAmount, err := strconv.ParseFloat(c.String("amount"), 64)
        if err != nil {
            return fmt.Errorf("Invalid deposit amount '%s': %w", c.String("amount"), err)
        }
        amount = depositAmount

    } else {

        // Get node status
        status, err := rp.NodeStatus()
        if err != nil {
            return err
        }

        // Get deposit amount options
        amountOptions := []string{
            "32 ETH (minipool begins staking immediately)",
            "16 ETH (minipool begins staking after ETH is assigned)",
        }
        if status.Trusted {
            amountOptions = append(amountOptions, "0 ETH  (minipool begins staking after ETH is assigned)")
        }

        // Prompt for amount
        selected, _ := cliutils.Select("Please choose an amount of ETH to deposit:", amountOptions)
        switch selected {
            case 0: amount = 32
            case 1: amount = 16
            case 2: amount = 0
        }

    }
    amountWei := eth.EthToWei(amount)

    // Check deposit can be made
    canDeposit, err := rp.CanNodeDeposit(amountWei)
    if err != nil {
        return err
    }
    if !canDeposit.CanDeposit {
        fmt.Println("Cannot make node deposit:")
        if canDeposit.InsufficientBalance {
            fmt.Println("The node's ETH balance is insufficient.")
        }
        if canDeposit.InvalidAmount {
            fmt.Println("The deposit amount is invalid.")
        }
        if canDeposit.DepositDisabled {
            fmt.Println("Node deposits are currently disabled.")
        }
        return nil
    }

    // Get minimum node fee
    var minNodeFee float64
    if c.String("min-fee") == "auto" {

        // Use suggested fee
        nodeFees, err := rp.NodeFee()
        if err != nil {
            return err
        }
        minNodeFee = nodeFees.SuggestedMinNodeFee

    } else if c.String("min-fee") != "" {

        // Parse fee
        minNodeFeePerc, err := strconv.ParseFloat(c.String("min-fee"), 64)
        if err != nil {
            return fmt.Errorf("Invalid minimum node fee '%s': %w", c.String("min-fee"), err)
        }
        minNodeFee = minNodeFeePerc / 100

    } else {

        // Prompt for fee
        nodeFees, err := rp.NodeFee()
        if err != nil {
            return err
        }
        minNodeFee = promptMinNodeFee(nodeFees.NodeFee, nodeFees.SuggestedMinNodeFee)

    }

    // Make deposit
    response, err := rp.NodeDeposit(amountWei, minNodeFee)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("The node deposit of %.6f ETH was made successfully.\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
    fmt.Printf("A new minipool was created at %s.\n", response.MinipoolAddress.Hex())
    return nil

}

