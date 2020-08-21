package service

import (
    "fmt"
    "strings"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
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

    // Prompt for confirmation
    response := cliutils.Prompt("Are you sure you want to pause the Rocket Pool service? Any staking minipools will be penalized! [y/n]", "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
    if strings.ToLower(response[:1]) == "n" {
        fmt.Println("Cancelled.")
        return nil
    }

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Pause service
    return rp.PauseService()

}


// Stop the Rocket Pool service
func stopService(c *cli.Context) error {

    // Prompt for confirmation
    response := cliutils.Prompt("Are you sure you want to stop the Rocket Pool service? Any staking minipools will be penalized, and ethereum nodes will lose sync progress! [y/n]", "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
    if strings.ToLower(response[:1]) == "n" {
        fmt.Println("Cancelled.")
        return nil
    }

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Stop service
    return rp.StopService()

}

