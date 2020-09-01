package service

import (
    "fmt"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/services"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Install the Rocket Pool service
func installService(c *cli.Context) error {

    // Get service details
    var location string
    if c.GlobalString("host") == "" {
        location = "locally"
    } else {
        location = fmt.Sprintf("at %s", c.GlobalString("host"))
    }

    // Prompt for confirmation
    if !cliutils.Confirm(fmt.Sprintf(
        "The Rocket Pool service will be installed %s --\nNetwork: %s\nVersion: %s\n\nAny existing configuration will be overwritten.\nAre you sure you want to continue?",
        location, c.String("network"), c.String("version"),
    )) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Install service
    return rp.InstallService(c.Bool("verbose"), c.Bool("use-wget"), c.Bool("ignore-deps"), c.String("network"), c.String("version"))

}


// View the Rocket Pool service status
func serviceStatus(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Print service status
    return rp.PrintServiceStatus()

}


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
    if !cliutils.Confirm("Are you sure you want to pause the Rocket Pool service? Any staking minipools will be penalized!") {
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
    if !cliutils.Confirm("Are you sure you want to stop the Rocket Pool service? Any staking minipools will be penalized, and ethereum nodes will lose sync progress!") {
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


// View the Rocket Pool service logs
func serviceLogs(c *cli.Context, serviceNames ...string) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Print service logs
    return rp.PrintServiceLogs(serviceNames...)

}


// View the Rocket Pool service stats
func serviceStats(c *cli.Context) error {

    // Get services
    rp, err := services.GetRocketPoolClient(c)
    if err != nil { return err }
    defer rp.Close()

    // Print service stats
    return rp.PrintServiceStats()

}

