package node

import (
    "fmt"

    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Config
const SuggestedNodeFeeDelta = -0.01 // 1% below current


func nodeDeposit(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get node status
    status, err := rp.NodeStatus()
    if err != nil {
        return err
    }

    var amount float64
    if !c.IsSet("amount") {
        // Get deposit amount options
        amountOptions := []string{
            "32 ETH (minipool begins staking immediately)",
            "16 ETH (minipool begins staking after ETH is assigned)",
        }
        if status.Trusted {
            amountOptions = append(amountOptions, "0 ETH  (minipool begins staking after ETH is assigned)")
        }
        // Prompt for eth amount
        selected, _ := cliutils.Select("Please choose an amount of ETH to deposit:", amountOptions)
        switch selected {
            case 0: amount = 32
            case 1: amount = 16
            case 2: amount = 0
        }
    } else {
        amount = c.Float64("amount")
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

    var minNodeFee float64 = 0
    if !c.IsSet("fee") && !c.IsSet("autofee") {
        // Get network node fees
        nodeFees, err := rp.NodeFee()
        if err != nil {
            return err
        }

        // Get suggested minimum node fee
        suggestedMinNodeFee := nodeFees.NodeFee + SuggestedNodeFeeDelta
        if suggestedMinNodeFee < nodeFees.MinNodeFee {
            suggestedMinNodeFee = nodeFees.MinNodeFee
        }

        // Prompt for minimum node fee
        minNodeFee = promptMinNodeFee(nodeFees.NodeFee, suggestedMinNodeFee)
    } else if c.IsSet("autofee") {
        // Get network node fees
        nodeFees, err := rp.NodeFee()
        if err != nil {
            return err
        }

        // Get suggested minimum node fee
        suggestedMinNodeFee := nodeFees.NodeFee + SuggestedNodeFeeDelta
        if suggestedMinNodeFee < nodeFees.MinNodeFee {
            suggestedMinNodeFee = nodeFees.MinNodeFee
        }
        minNodeFee = suggestedMinNodeFee
    } else {
        // fee is set
        minNodeFee = c.Float64("fee") / 100
        if minNodeFee < 0 || minNodeFee > 1 {
            return fmt.Errorf("Invalid commission rate: %.5f", minNodeFee)
        }
    }

    //fmt.Printf("Amount: %.10f ETH, Fee: %.5f %%\n", eth.WeiToEth(amountWei), minNodeFee * 100)

    // Make deposit
    response, err := rp.NodeDeposit(amountWei, minNodeFee)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("The node deposit of %.2f ETH was made successfully.\n", eth.WeiToEth(amountWei))
    fmt.Printf("A new minipool was created at %s.\n", response.MinipoolAddress.Hex())
    return nil

}
