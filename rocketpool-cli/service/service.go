package service

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

// Install the Rocket Pool service
func installService(c *cli.Context) error {

    // Get install location
    var location string
    if c.GlobalString("host") == "" {
        location = "locally"
    } else {
        location = fmt.Sprintf("at %s", c.GlobalString("host"))
    }

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(fmt.Sprintf(
        "The Rocket Pool service will be installed %s --\nNetwork: %s\nVersion: %s\n\nAny existing configuration will be overwritten.\nAre you sure you want to continue?",
        location, c.String("network"), c.String("version"),
    ))) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Install service
    err = rp.InstallService(c.Bool("verbose"), c.Bool("no-deps"), c.String("network"), c.String("version"))
    if err != nil { return err }

    // Print success message & return
    colorReset := "\033[0m"
    colorYellow := "\033[33m"
    fmt.Println("")
    fmt.Printf("The Rocket Pool service was successfully installed %s!\n", location)
    if c.GlobalString("host") == "" {
        fmt.Println("")
        fmt.Printf("%sNOTE:\nIf this is your first time installing Rocket Pool, please start a new shell session by logging out and back in or restarting the machine.\n", colorYellow)
        fmt.Println("This is necessary for your user account to have permissions to use Docker.")
        fmt.Printf("If you have installed Rocket Pool previously and are just upgrading, you can safely ignore this message.%s\n", colorReset)
        fmt.Println("")
    }
    fmt.Println("Please run 'rocketpool service config' to configure the service before starting it.")
    return nil

}


// Install the Rocket Pool update tracker for the metrics dashboard
func installUpdateTracker(c *cli.Context) error {

    // Get install location
    var location string
    if c.GlobalString("host") == "" {
        location = "locally"
    } else {
        location = fmt.Sprintf("at %s", c.GlobalString("host"))
    }

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm(
        "This will add the ability to display any available Operating System updates or new Rocket Pool versions on the metrics dashboard. " +
        "Are you sure you want to install the update tracker?")) {
            fmt.Println("Cancelled.")
            return nil
    }

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Install service
    err = rp.InstallUpdateTracker(c.Bool("verbose"), c.String("version"))
    if err != nil { return err }

    // Print success message & return
    colorReset := "\033[0m"
    colorYellow := "\033[33m"
    fmt.Println("")
    fmt.Printf("The Rocket Pool update tracker service was successfully installed %s!\n", location)
    if c.GlobalString("host") == "" {
        fmt.Println("")
        fmt.Printf("%sNOTE:\nPlease run 'docker restart rocketpool_exporter' to enable update tracking on the metrics dashboard.%s\n", colorYellow, colorReset)
        fmt.Println("")
    }
    return nil

}


// View the Rocket Pool service status
func serviceStatus(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Print what network we're on the network
    err = cliutils.PrintNetwork(rp)
    if err != nil {
        return err
    }

    // Print service status
    return rp.PrintServiceStatus(getComposeFiles(c))

}


// Start the Rocket Pool service
func startService(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Update the Prometheus template with the assigned ports
    userConfig, err := rp.LoadUserConfig()
    if err != nil {
        return fmt.Errorf("Error loading user settings: %w", err)
    }
    metricsEnabled := userConfig.Metrics.Enabled
    if metricsEnabled {
        err := rp.UpdatePrometheusConfiguration(userConfig.Metrics.Settings)
        if err != nil {
            return err
        }
    }

    // Start service
    return rp.StartService(getComposeFiles(c))

}


// Pause the Rocket Pool service
func pauseService(c *cli.Context) error {

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to pause the Rocket Pool service? Any staking minipools will be penalized!")) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Pause service
    return rp.PauseService(getComposeFiles(c))

}


// Stop the Rocket Pool service
func stopService(c *cli.Context) error {

    // Prompt for confirmation
    if !(c.Bool("yes") || cliutils.Confirm("Are you sure you want to terminate the Rocket Pool service? Any staking minipools will be penalized, chain databases will be deleted, and ethereum nodes will lose ALL sync progress!")) {
        fmt.Println("Cancelled.")
        return nil
    }

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Stop service
    return rp.StopService(getComposeFiles(c))

}


// View the Rocket Pool service logs
func serviceLogs(c *cli.Context, serviceNames ...string) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Print service logs
    return rp.PrintServiceLogs(getComposeFiles(c), c.String("tail"), serviceNames...)

}


// View the Rocket Pool service stats
func serviceStats(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Print service stats
    return rp.PrintServiceStats(getComposeFiles(c))

}


// View the Rocket Pool service version information
func serviceVersion(c *cli.Context) error {

    // Get RP client
    rp, err := rocketpool.NewClientFromCtx(c)
    if err != nil { return err }
    defer rp.Close()

    // Get RP service version
    serviceVersion, err := rp.GetServiceVersion()
    if err != nil { return err }

    // Get config
    cfg, err := rp.LoadMergedConfig()
    if err != nil { return err }
    eth1Client := cfg.GetSelectedEth1Client()
    eth2Client := cfg.GetSelectedEth2Client()

    // Get client versions
    var eth1ClientVersion string
    var eth2ClientVersion string
    var eth2ClientImage string
    if eth1Client != nil {
        eth1ClientVersion = fmt.Sprintf("%s (%s)", eth1Client.Name, eth1Client.Image)
    } else {
        eth1ClientVersion = "(none)"
    }
    if eth2Client != nil {
        if eth2Client.Image != "" {
            eth2ClientImage = eth2Client.Image
        } else {
            eth2ClientImage = eth2Client.BeaconImage
        }
        eth2ClientVersion = fmt.Sprintf("%s (%s)", eth2Client.Name, eth2ClientImage)
    } else {
        eth2ClientVersion = "(none)"
    }

    // Print version info
    fmt.Printf("Rocket Pool client version: %s\n", c.App.Version)
    fmt.Printf("Rocket Pool service version: %s\n", serviceVersion)
    fmt.Printf("Selected Eth 1.0 client: %s\n", eth1ClientVersion)
    fmt.Printf("Selected Eth 2.0 client: %s\n", eth2ClientVersion)
    return nil

}


// Get the compose file paths for a CLI context
func getComposeFiles(c *cli.Context) []string {
    return c.Parent().StringSlice("compose-file")
}

