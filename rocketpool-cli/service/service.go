package service

import (
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
)


// Start the Rocket Pool service
func startService(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Start service
    return rp.StartService()

}


// Pause the Rocket Pool service
func pauseService(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Pause service
    return rp.PauseService()

}


// Stop the Rocket Pool service
func stopService(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Stop service
    return rp.StopService()

}

