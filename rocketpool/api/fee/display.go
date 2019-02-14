package fee

import (
    "errors"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Display the current user fee
func displayUserFee(c *cli.Context) error {

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

    // Get user fee
    userFee := new(*big.Int)
    err = rp.Contracts["rocketNodeSettings"].Call(nil, userFee, "getFeePerc")
    if err != nil {
        return errors.New("Error retrieving node user fee percentage setting: " + err.Error())
    }
    userFeePerc := eth.WeiToEth(*userFee) * 100

    // Log & return
    fmt.Println(fmt.Sprintf("The current Rocket Pool user fee paid to node operators is %.2f%% of rewards", userFeePerc))
    return nil

}

