package deposit

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/deposit"
    "github.com/rocket-pool/smartnode/shared/services"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
)


// Get the current deposit RPL requirement
func getRplRequired(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        CM: true,
        LoadContracts: []string{"rocketMinipoolSettings", "rocketNodeAPI", "rocketPool"},
        WaitClientConn: true,
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get RPL requirement
    required, err := deposit.GetRplRequired(p)
    if err != nil { return err }

    // Print output & return
    for _, duration := range required.Durations {
        fmt.Fprintln(p.Output, fmt.Sprintf(
            "Depositing %.2f ETH for %s requires %.2f RPL @ %.2f RPL / ETH. Current network utilisation for %s is %.2f%%.",
            eth.WeiToEth(duration.EtherAmountWei),
            duration.DurationId,
            eth.WeiToEth(duration.RplAmountWei),
            eth.WeiToEth(duration.RplRatioWei),
            duration.DurationId,
            duration.NetworkUtilisation * 100))
    }
    return nil

}

