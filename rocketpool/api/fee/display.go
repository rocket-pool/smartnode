package fee

import (
    "errors"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/database"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Display the current user fee
func displayUserFee(c *cli.Context) error {

    // Initialise database
    db, err := database.NewDatabase(c.GlobalString("database"))
    if err != nil {
        return err
    }
    defer db.Close()

    // Connect to ethereum node
    client, err := ethclient.Dial(c.GlobalString("provider"))
    if err != nil {
        return errors.New("Error connecting to ethereum node: " + err.Error())
    }

    // Initialise Rocket Pool contract manager
    rp, err := rocketpool.NewContractManager(client, c.GlobalString("storageAddress"))
    if err != nil {
        return err
    }

    // Load Rocket Pool contracts
    err = rp.LoadContracts([]string{"rocketNodeSettings"})
    if err != nil {
        return err
    }

    // Get current user fee
    userFee := new(*big.Int)
    err = rp.Contracts["rocketNodeSettings"].Call(nil, userFee, "getFeePerc")
    if err != nil {
        return errors.New("Error retrieving node user fee percentage setting: " + err.Error())
    }
    userFeePerc := eth.WeiToEth(*userFee) * 100

    // Get target user fee
    targetUserFeePerc := new(float64)
    err = db.Get("user.fee.target", targetUserFeePerc)
    if err != nil {
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

