package faucet

import (
    "fmt"
    "strings"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func faucetWithdraw(c *cli.Context, token string) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Withdraw from faucet
    if _, err := rp.FaucetWithdraw(token); err != nil {
        return err
    }

    // Log & return
    fmt.Printf("Successfully withdrew %s from the faucet. Run 'rocketpool node status' to check your balance.", strings.ToUpper(token))
    return nil

}

