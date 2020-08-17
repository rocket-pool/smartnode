package config

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


func configureService(c *cli.Context) error {

    // Initialize RP client
    rp, err := rocketpool.NewClient(c.GlobalString("host"), c.GlobalString("user"), c.GlobalString("key"))
    if err != nil {
        return err
    }
    defer rp.Close()

    // Load global config
    globalConfig, err := rp.LoadGlobalConfig()
    if err != nil {
        return err
    }
    _ = globalConfig

    // Return
    return nil

}

