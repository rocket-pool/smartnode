package network

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func challenge(c *cli.Context, address common.Address) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get node fee
    _, err = rp.Challenge(address)
    if err != nil {
        return err
    }

    // Print & return
    fmt.Println("Done.")
    return nil

}

