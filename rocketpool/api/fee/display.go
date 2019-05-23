package fee

import (
    "errors"
    "fmt"
    "math/big"

    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Display the current user fee
func displayUserFee(c *cli.Context) error {

    // Initialise services
    p, err := services.NewProvider(c, services.ProviderOpts{
        DB: true,
        CM: true,
        LoadContracts: []string{"rocketNodeSettings"},
    })
    if err != nil {
        return err 
    }

    // Open database
    if err := p.DB.Open(); err != nil {
        return err
    }
    defer p.DB.Close()

    // Get current user fee
    userFee := new(*big.Int)
    if err := p.CM.Contracts["rocketNodeSettings"].Call(nil, userFee, "getFeePerc"); err != nil {
        return errors.New("Error retrieving node user fee percentage setting: " + err.Error())
    }
    userFeePerc := eth.WeiToEth(*userFee) * 100

    // Get target user fee
    targetUserFeePerc := new(float64)
    if err := p.DB.Get("user.fee.target", targetUserFeePerc); err != nil {
        *targetUserFeePerc = -1
    }

    // Log & return
    fmt.Println(fmt.Sprintf("The current Rocket Pool user fee paid to node operators is %.2f%% of rewards", userFeePerc))
    if *targetUserFeePerc == -1 {
        fmt.Println("The target Rocket Pool user fee to vote for is not set")
    } else {
        fmt.Println(fmt.Sprintf("The target Rocket Pool user fee to vote for is %.2f%% of rewards", *targetUserFeePerc))
    }
    return nil

}

