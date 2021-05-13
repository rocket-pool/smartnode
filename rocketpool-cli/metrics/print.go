package metrics

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func print(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Print metrics
    return rp.PrintMetricsOutput()

}

