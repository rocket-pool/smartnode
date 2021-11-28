package service

import (
	"fmt"
	"regexp"
	"time"

	"github.com/urfave/cli"

	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)

const dockerImageRegex string = ".*/(?P<image>.*):.*"
const colorReset string = "\033[0m"
const colorRed string = "\033[31m"
const colorYellow string = "\033[33m"
const clearLine string = "\033[2K"

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
    err = rp.InstallService(c.Bool("verbose"), c.Bool("no-deps"), c.String("network"), c.String("version"), c.String("path"))
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

    // Print what network we're on
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

    if !c.Bool("ignore-slash-timer") {
        // Do the client swap check
        err := checkForValidatorChange(rp, userConfig)
        if err != nil {
            fmt.Printf("%sWarning: couldn't verify that the validator container can be safely restarted:\n\t%s\n", colorYellow, err.Error())
            fmt.Println("If you are changing to a different ETH2 client, it may resubmit an attestation you have already submitted.")
            fmt.Println("This will slash your validator!")
            fmt.Println("To prevent slashing, you must wait 15 minutes from the time you stopped the clients before starting them again.\n")
            fmt.Println("**If you did NOT change clients, you can safely ignore this warning.**\n")
            if !cliutils.Confirm(fmt.Sprintf("Press y when you understand the above warning, have waited, and are ready to start Rocket Pool:%s", colorReset)) {
                fmt.Println("Cancelled.")
                return nil
            }
        }
    } else {
        fmt.Printf("%sIgnoring anti-slashing safety delay.%s\n", colorYellow, colorReset)
    }

    // Start service
    return rp.StartService(getComposeFiles(c))

}


func checkForValidatorChange(rp *rocketpool.Client, userConfig config.RocketPoolConfig) (error) {

    // Get the current validator client
    currentValidatorImageString, err := rp.GetDockerImage("rocketpool_validator")
    currentValidatorName, err := getDockerImageName(currentValidatorImageString)
    if err != nil {
        return fmt.Errorf("Error getting current validator image name: %w", err)
    }

    // Get the new validator client according to the settings file
    globalConfig, err := rp.LoadGlobalConfig()
    if err != nil {
        return fmt.Errorf("Error loading global settings: %w", err)
    }
    newClient := globalConfig.Chains.Eth2.GetClientById(userConfig.Chains.Eth2.Client.Selected)
    if newClient == nil {
        return fmt.Errorf("Error getting selected client - either it does not exist (user has not run `rocketpool service config` yet) or the selected client is invalid.")
    }
    pendingValidatorImageString := newClient.GetValidatorImage()
    pendingValidatorName, err := getDockerImageName(pendingValidatorImageString)
    if err != nil {
        return fmt.Errorf("Error getting pending validator image name: %w", err)
    }

    // Compare the clients and warn if necessary
    if currentValidatorName == pendingValidatorName {
        fmt.Printf("Validator client [%s] was previously used - no slashing prevention delay necessary.\n", currentValidatorName)
    } else if currentValidatorName == "" {
        fmt.Println("This is the first time starting Rocket Pool - no slashing prevention delay necessary.")
    } else {

        // Get the time that the container responsible for validator duties exited
        validatorDutyContainerName := getContainerNameForValidatorDuties(currentValidatorName, rp)
        validatorFinishTime, err := rp.GetDockerContainerShutdownTime(validatorDutyContainerName)
        if err != nil {
            return fmt.Errorf("Error getting validator shutdown time: %w", err)
        }

        // If it hasn't exited yet, shut it down
        zeroTime := time.Time{}
        status, err := rp.GetDockerStatus(validatorDutyContainerName)
        if err != nil {
            return fmt.Errorf("Error getting container [%s] status: %w", validatorDutyContainerName, err)
        }
        if validatorFinishTime == zeroTime || status == "running" {
            fmt.Printf("%sValidator is currently running, stopping it...%s\n", colorYellow, colorReset)
            response, err := rp.StopContainer(validatorDutyContainerName)
            validatorFinishTime = time.Now()
            if err != nil {
                return fmt.Errorf("Error stopping container [%s]: %w", validatorDutyContainerName, err)
            }
            if response != validatorDutyContainerName {
                return fmt.Errorf("Unexpected response when stopping container [%s]: %s", validatorDutyContainerName, response)
            }
        }

        // Print the warning and start the time lockout
        safeStartTime := validatorFinishTime.Add(15 * time.Minute)
        remainingTime := time.Until(safeStartTime)
        if remainingTime <= 0 {
            fmt.Printf("The validator has been offline for %s, which is long enough to prevent slashing.\n", time.Since(validatorFinishTime))
            fmt.Println("The new client can be safely started.")
        } else {
            fmt.Printf("%s=== WARNING ===\n", colorRed)
            fmt.Printf("You have changed your validator client from %s to %s.\n", currentValidatorName, pendingValidatorName)
            fmt.Println("If you have active validators, starting the new client immediately will cause them to be slashed due to duplicate attestations!")
            fmt.Println("To prevent slashing, Rocket Pool will delay activating the new client for 15 minutes.")
            fmt.Printf("If you want to bypass this cooldown and understand the risks, rerun this command with the `--ignore-slash-timer` flag.%s\n\n", colorReset)

            // Wait for 15 minutes
            for remainingTime > 0 {
                fmt.Printf("Remaining time: %s", remainingTime)
                time.Sleep(1 * time.Second)
                remainingTime = time.Until(safeStartTime)
                fmt.Printf("%s\r", clearLine)
            }

            fmt.Println(colorReset)
            fmt.Println("You may now safely start the validator without fear of being slashed.")
        }
    }

    return nil
}


// Get the name of the container responsible for validator duties based on the client name
// TODO: this is temporary and can change, clean it up when Nimbus supports split mode
func getContainerNameForValidatorDuties(CurrentValidatorClientName string, rp *rocketpool.Client) (string) {

    if (CurrentValidatorClientName == "nimbus") {
        return "rocketpool_eth2"
    } else {
        return "rocketpool_validator"
    }

}


// Get the time that the container responsible for validator duties exited
func getValidatorFinishTime(CurrentValidatorClientName string, rp *rocketpool.Client) (time.Time, error) {

    var validatorFinishTime time.Time
    var err error
    if (CurrentValidatorClientName == "nimbus") {
        validatorFinishTime, err = rp.GetDockerContainerShutdownTime("rocketpool_eth2")
    } else {
        validatorFinishTime, err = rp.GetDockerContainerShutdownTime("rocketpool_validator")
    }

    return validatorFinishTime, err

}


// Extract the image name from a Docker image string
func getDockerImageName(imageString string) (string, error) {

    // Return the empty string if the validator didn't exist (probably because this is the first time starting it up)
    if imageString == "" {
        return "", nil
    }

    reg := regexp.MustCompile(dockerImageRegex)
    matches := reg.FindStringSubmatch(imageString)
    if matches == nil {
        return "", fmt.Errorf("Couldn't parse the Docker image string [%s]", imageString)
    }
    imageIndex := reg.SubexpIndex("image")
    if imageIndex == -1 {
        return "", fmt.Errorf("Image name not found in Docker image [%s]", imageString)
    }

    imageName := matches[imageIndex]
    return imageName, nil
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

    // Print what network we're on
    err = cliutils.PrintNetwork(rp)
    if err != nil {
        return err
    }

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

