package config

import (
    "os"

    "github.com/rocket-pool/smartnode/shared/utils/config"
)


// Configure the Rocket Pool service
func configureService() error {

    // Load config
    rpConfig, err := config.Load(os.Getenv("RP_PATH"))
    if err != nil { return err }
    _ = rpConfig

    // Return
    return nil

}

