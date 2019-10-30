package deposit

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/deposit"
    "github.com/rocket-pool/smartnode/shared/services"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Make a deposit
func makeDeposit(c *cli.Context, durationId string) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        AM: true,
        KM: true,
        Client: true,
        CM: true,
        NodeContractAddress: true,
        NodeContract: true,
        LoadContracts: []string{"rocketDepositQueue", "rocketETHToken", "rocketMinipoolSettings", "rocketNodeAPI", "rocketNodeSettings", "rocketPool", "rocketPoolToken"},
        LoadAbis: []string{"rocketNodeContract"},
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get deposit status
    status, err := deposit.GetDepositStatus(p)
    if err != nil { return err }

    // Reserve deposit if reservation doesn't exist
    var statusFormat string
    if status.ReservationExists {
        statusFormat = "Node already has a deposit reservation requiring %.2f ETH and %.2f RPL, with a staking duration of %s and expiring at %s"
    } else {
        statusFormat = "Deposit reservation made successfully, requiring %.2f ETH and %.2f RPL, with a staking duration of %s and expiring at %s"

        // Generate new validator key
        validatorKey, err := p.KM.CreateValidatorKey()
        if err != nil { return err }

        // Check node deposit can be reserved
        reserved, err := deposit.CanReserveDeposit(p, validatorKey)
        if err != nil { return err }

        // Check response
        if reserved.DepositsDisabled {
            fmt.Fprintln(p.Output, "Node deposits are currently disabled in Rocket Pool")
            return nil
        }
        if reserved.PubkeyUsed {
            fmt.Fprintln(p.Output, "The validator public key is already in use")
            return nil
        }

        // Reserve deposit
        reserved, err = deposit.ReserveDeposit(p, validatorKey, durationId)
        if err != nil { return err }

        // Get deposit status
        status, err = deposit.GetDepositStatus(p)
        if err != nil { return err }

    }

    // Print deposit status
    fmt.Fprintln(p.Output, fmt.Sprintf(
        statusFormat,
        eth.WeiToEth(status.ReservationEtherRequiredWei),
        eth.WeiToEth(status.ReservationRplRequiredWei),
        status.ReservationStakingDurationID,
        status.ReservationExpiryTime.Format("2006-01-02, 15:04 -0700 MST")))
    fmt.Fprintln(p.Output, fmt.Sprintf("Node deposit contract has a balance of %.2f ETH and %.2f RPL", eth.WeiToEth(status.NodeBalanceEtherWei), eth.WeiToEth(status.NodeBalanceRplWei)))

    // Prompt for action
    response := cliutils.Prompt(p.Input, p.Output, "Would you like to:\n1. Complete the deposit;\n2. Cancel the deposit; or\n3. Finish later?", "^(1|2|3)$", "Please answer '1', '2' or '3'")
    switch response {

        // Complete deposit
        case "1":

            // Check deposit can be completed
            completed, err := deposit.CanCompleteDeposit(p)
            if err != nil { return err }

            // Check response
            // TODO: fix insufficient balance messages
            if completed.DepositsDisabled {
                fmt.Fprintln(p.Output, "Node deposits are currently disabled in Rocket Pool")
                return nil
            }
            if completed.MinipoolCreationDisabled {
                fmt.Fprintln(p.Output, "Minipool creation is currently disabled in Rocket Pool")
                return nil
            }
            if completed.InsufficientNodeEtherBalance {
                fmt.Fprintln(p.Output, fmt.Sprintf("Node balance of %.2f ETH plus account balance of %.2f ETH is not enough to cover requirement of %.2f ETH"))
                return nil
            }
            if completed.InsufficientNodeRplBalance {
                fmt.Fprintln(p.Output, fmt.Sprintf("Node balance of %.2f RPL is not enough to cover requirement of %.2f RPL"))
                return nil
            }

            // Confirm ETH send
            // TODO: implement

            // Confirm & perform RPL send
            // TODO: implement

            // Complete
            completed, err = deposit.CompleteDeposit(p, completed.EtherRequiredWei, completed.DepositDurationId)
            if err != nil { return err }

            // Print output
            if completed.Success {
                fmt.Fprintln(p.Output, "Deposit completed successfully, minipool created at", completed.MinipoolAddress.Hex())
            }

        // Cancel deposit
        case "2":

            // Cancel
            cancelled, err := deposit.CancelDeposit(p)
            if err != nil { return err }

            // Print output
            if cancelled.Success {
                fmt.Fprintln(p.Output, "Deposit reservation cancelled successfully")
            }

    }

    // Return
    return nil

}

