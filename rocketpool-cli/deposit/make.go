package deposit

import (
    "fmt"

    "github.com/urfave/cli"

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
        WaitClientConn: true,
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get deposit status
    status, err := getDepositStatus(p)
    if err != nil { return err }

    // Reserve deposit if reservation doesn't exist
    var statusFormat string
    if status.Reservation.Exists {
        statusFormat = "Node already has a deposit reservation requiring %.2f ETH and %.2f RPL, with a staking duration of %s and expiring at %s"
    } else {
        statusFormat = "Deposit reservation made successfully, requiring %.2f ETH and %.2f RPL, with a staking duration of %s and expiring at %s"

        // Reserve deposit
        if err := reserveDeposit(p, durationId); err != nil { return err }

        // Get deposit status
        status, err = getDepositStatus(p)
        if err != nil { return err }

    }

    // Display deposit status
    fmt.Fprintln(p.Output, fmt.Sprintf(
        statusFormat,
        eth.WeiToEth(status.Reservation.EtherRequiredWei),
        eth.WeiToEth(status.Reservation.RplRequiredWei),
        status.Reservation.StakingDurationID,
        status.Reservation.ExpiryTime.Format("2006-01-02, 15:04 -0700 MST")))
    fmt.Fprintln(p.Output, fmt.Sprintf("Node deposit contract has a balance of %.2f ETH and %.2f RPL", eth.WeiToEth(status.Balances.EtherWei), eth.WeiToEth(status.Balances.RplWei)))

    // Prompt for action
    response := cliutils.Prompt(p.Input, p.Output, "Would you like to:\n1. Complete the deposit;\n2. Cancel the deposit; or\n3. Finish later?", "^(1|2|3)$", "Please answer '1', '2' or '3'")
    switch response {

        // Complete deposit
        case "1":
            if minipoolCreated, err := completeDeposit(p); err != nil {
                return err
            } else if minipoolCreated != nil {
                fmt.Fprintln(p.Output, "Deposit completed successfully, minipool created at", minipoolCreated.Address.Hex())
            }

        // Cancel deposit
        case "2":
            if err := cancelDeposit(p); err != nil {
                return err
            } else {
                fmt.Fprintln(p.Output, "Deposit reservation cancelled successfully")
            }

    }

    // Return
    return nil

}

