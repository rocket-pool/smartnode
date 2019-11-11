package fee

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/api/fee"
    "github.com/rocket-pool/smartnode/shared/services"
)


// Display the current user fee
func displayUserFee(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        DB: true,
        CM: true,
        LoadContracts: []string{"rocketNodeSettings"},
        WaitClientConn: true,
        WaitClientSync: true,
        WaitRocketStorage: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Get user fee
    userFee, err := fee.GetUserFee(p)
    if err != nil { return err }

    // Print output & return
    fmt.Fprintln(p.Output, fmt.Sprintf("The current Rocket Pool user fee paid to node operators is %.2f%% of rewards", userFee.CurrentUserFeePerc))
    if userFee.TargetUserFeePerc == -1 {
        fmt.Fprintln(p.Output, "The target Rocket Pool user fee to vote for is not set")
    } else {
        fmt.Fprintln(p.Output, fmt.Sprintf("The target Rocket Pool user fee to vote for is %.2f%% of rewards", userFee.TargetUserFeePerc))
    }
    return nil

}

