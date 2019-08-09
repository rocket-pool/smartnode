package fee

import (
    "errors"
    "fmt"

    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Set the target user fee to vote for
func setTargetUserFee(c *cli.Context, feePercent float64) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        DB: true,
    })
    if err != nil { return err }
    defer p.Cleanup()

    // Open database
    if err := p.DB.Open(); err != nil {
        return err
    }
    defer p.DB.Close()

    // Set target user fee percent
    if err := p.DB.Put("user.fee.target", feePercent); err != nil {
        return errors.New("Error setting target user fee percentage: " + err.Error())
    }

    // Log & return
    fmt.Fprintln(p.Output, fmt.Sprintf("Target user fee to vote for successfully set to %.2f%%", feePercent))
    return nil

}

