package network

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
)


func getNodeFee(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Get node fee
    response, err := rp.NodeFee()
    if err != nil {
        return err
    }

    // Print & return
    fmt.Printf("The current network node fee is %f%%\n", response.NodeFee * 100)
    return nil

}

