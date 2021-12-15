package node

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/gas"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
	"github.com/rocket-pool/smartnode/shared/utils/math"
)

// Config
const DefaultMaxNodeFeeSlippage = 0.01 // 1% below current network fee


func nodeDeposit(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Make sure ETH2 is on the correct chain
    depositContractInfo, err := rp.DepositContractInfo()
    if err != nil {
        return err
    }
    if depositContractInfo.RPNetwork != depositContractInfo.BeaconNetwork ||
       depositContractInfo.RPDepositContract != depositContractInfo.BeaconDepositContract {
        cliutils.PrintDepositMismatchError(
            depositContractInfo.RPNetwork,
            depositContractInfo.BeaconNetwork,
            depositContractInfo.RPDepositContract,
            depositContractInfo.BeaconDepositContract)
        return nil
    } else {
        fmt.Println("Your eth2 client is on the correct network.\n")
    }

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

    // Get network node fees
    nodeFees, err := rp.NodeFee()
    if err != nil {
        return err
    }

    // Get minimum node fee
    var minNodeFee float64
    if c.String("max-slippage") == "auto" {

        // Use default max slippage
        minNodeFee = nodeFees.NodeFee - DefaultMaxNodeFeeSlippage
        if minNodeFee < nodeFees.MinNodeFee { minNodeFee = nodeFees.MinNodeFee }

    } else if c.String("max-slippage") != "" {

        // Parse max slippage
        maxNodeFeeSlippagePerc, err := strconv.ParseFloat(c.String("max-slippage"), 64)
        if err != nil {
            return fmt.Errorf("Invalid maximum commission rate slippage '%s': %w", c.String("max-slippage"), err)
        }
        maxNodeFeeSlippage := maxNodeFeeSlippagePerc / 100

        // Calculate min node fee
        minNodeFee = nodeFees.NodeFee - maxNodeFeeSlippage
        if minNodeFee < nodeFees.MinNodeFee { minNodeFee = nodeFees.MinNodeFee }

    } else {

        // Prompt for min node fee
        minNodeFee = promptMinNodeFee(nodeFees.NodeFee, nodeFees.MinNodeFee)

    }

    // Get minipool salt
    var salt *big.Int
    if c.String("salt") != "" {
        var success bool
        salt, success = big.NewInt(0).SetString(c.String("salt"), 0)
        if !success {
            return fmt.Errorf("Invalid minipool salt: %s", c.String("salt"))
        }
    } else {
        buffer := make([]byte, 32)
        _, err = rand.Read(buffer)
        if err != nil {
            return fmt.Errorf("Error generating random salt: %w", err)
        }
        salt = big.NewInt(0).SetBytes(buffer)
    }

    // Check deposit can be made
    canDeposit, err := rp.CanNodeDeposit(amountWei, minNodeFee, salt)
    if err != nil {
        return err
    }
    if !canDeposit.CanDeposit {
        fmt.Println("Cannot make node deposit:")
        if canDeposit.InsufficientBalance {
            fmt.Println("The node's ETH balance is insufficient.")
        }
        if canDeposit.InsufficientRplStake {
            fmt.Println("The node has not staked enough RPL to collateralize a new minipool.")
        }
        if canDeposit.InvalidAmount {
            fmt.Println("The deposit amount is invalid.")
        }
        if canDeposit.UnbondedMinipoolsAtMax {
            fmt.Println("The node cannot create any more unbonded minipools.")
        }
        if canDeposit.DepositDisabled {
            fmt.Println("Node deposits are currently disabled.")
        }
        if !canDeposit.InConsensus {
            fmt.Println("The RPL price and total effective staked RPL of the network are still being voted on by the Oracle DAO.\nPlease try again in a few minutes.")
        }
        return nil
    }

    if c.String("salt") != "" {
        fmt.Printf("Using custom salt %s, your minipool address will be %s.\n\n", c.String("salt"), canDeposit.MinipoolAddress.Hex())
    }

    // Check to see if eth2 is synced
    colorReset := "\033[0m"
    colorRed := "\033[31m"
    colorYellow := "\033[33m"
    syncResponse, err := rp.NodeSync()
    if err != nil {
        fmt.Printf("%s**WARNING**: Can't verify the sync status of your eth2 client.\nYOU WILL LOSE ETH if your minipool is activated before it is fully synced.\n" +
        "Reason: %s\n%s", colorRed, err, colorReset)
    } else {
        if !syncResponse.Eth2Synced {
            fmt.Printf("%s**WARNING**: your eth2 client is still syncing.\nYOU WILL LOSE ETH if your minipool is activated before it is fully synced.\n%s", colorRed, colorReset)
        } else {
            fmt.Printf("Your eth2 client is synced, you may safely create a minipool.\n")
        }
    }

    // Assign max fees
    err = gas.AssignMaxFeeAndLimit(canDeposit.GasInfo, rp, c.Bool("yes"))
    if err != nil{
        return err
    }

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf(
        "You are about to deposit %.6f ETH to create a minipool with a minimum possible commission rate of %f%%.\n" + 
        "%sARE YOU SURE YOU WANT TO DO THIS? Running a minipool is a long-term commitment, and this action cannot be undone!%s",
        math.RoundDown(eth.WeiToEth(amountWei), 6),
        minNodeFee * 100,
        colorYellow,
        colorReset))) {
            fmt.Println("Cancelled.")
            return nil
    }

    // Make deposit
    response, err := rp.NodeDeposit(amountWei, minNodeFee, salt)
    if err != nil {
        return err
    }

    // Log and wait for the minipool address
    fmt.Printf("Creating minipool...\n")
    cliutils.PrintTransactionHash(rp, response.TxHash)
    _, err = rp.WaitForTransaction(response.TxHash)
    if err != nil {
        return err
    }

    // Log & return
    fmt.Printf("The node deposit of %.6f ETH was made successfully!\n", math.RoundDown(eth.WeiToEth(amountWei), 6))
    fmt.Printf("Your new minipool's address is: %s\n", response.MinipoolAddress)
    fmt.Printf("The validator pubkey is: %s\n\n", response.ValidatorPubkey.Hex())

    fmt.Println("Your minipool is now in Initialized status.")
    fmt.Println("Once the 16 ETH deposit has been matched by the staking pool, it will move to Prelaunch status.")
    fmt.Printf("After that, it will move to Staking status once %s have passed.\n", response.ScrubPeriod)
    fmt.Println("You can watch its progress using `rocketpool service logs node`.")

    return nil

}

