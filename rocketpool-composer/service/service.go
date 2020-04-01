package service

import (
    "bufio"
    "errors"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
    "github.com/rocket-pool/smartnode/shared/utils/config"
)


// Start Rocket Pool service
func startService() error {
    cmd, err := compose("up", "-d")
    if err != nil { return err }
    return printOutput(cmd)
}


// Pause Rocket Pool service
func pauseService() error {

    // Prompt for confirmation
    response := cliutils.Prompt(nil, nil, "Are you sure you want to pause the Rocket Pool services? Any staking minipools will be penalized! [y/n]", "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
    if strings.ToLower(response[:1]) == "n" { return nil }

    // Pause service
    cmd, err := compose("stop")
    if err != nil { return err }
    return printOutput(cmd)

}


// Stop Rocket Pool service
func stopService() error {

    // Prompt for confirmation
    response := cliutils.Prompt(nil, nil, "Are you sure you want to stop the Rocket Pool services? Any staking minipools will be penalized, and ethereum nodes will lose sync progress! [y/n]", "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
    if strings.ToLower(response[:1]) == "n" { return nil }

    // Stop service
    cmd, err := compose("down", "-v")
    if err != nil { return err }
    return printOutput(cmd)

}


// Scale Rocket Pool service
func scaleService(args ...string) error {
    cmd, err := compose(append([]string{"scale"}, args...)...)
    if err != nil { return err }
    return printOutput(cmd)
}


// Print Rocket Pool service logs
func serviceLogs(args ...string) error {
    cmd, err := compose(append([]string{"logs", "-f"}, args...)...)
    if err != nil { return err }
    return printOutput(cmd)
}


// Print Rocket Pool service resource stats
func serviceStats() error {

    // Get service container IDs
    cmd, err := compose("ps", "-q")
    if err != nil { return err }
    containerIds, err := readOutputLines(cmd)
    if err != nil { return err }

    // Print stats
    return printOutput(exec.Command("docker", append([]string{"stats"}, containerIds...)...))

}


// Execute a Rocket Pool CLI command
func execCommand(args ...string) error {
    cmd, err := compose(append([]string{"exec", "-T", "cli", "/go/bin/rocketpool", "run"}, args...)...)
    if err != nil { return err }
    return printOutput(cmd)
}


// Build a docker-compose command with the given arguments
func compose(args ...string) (*exec.Cmd, error) {

    // Get RP_PATH
    rpPath := os.Getenv("RP_PATH")

    // Load RP config
    rpConfig, err := config.Load(rpPath)
    if err != nil { return nil, err }

    // Check config
    if rpConfig.Chains.Eth1.Client.Selected == "" {
        return nil, errors.New("No Eth1 client selected. Please run 'rocketpool config' and try again.")
    }
    if rpConfig.Chains.Eth2.Client.Selected == "" {
        return nil, errors.New("No Eth2 client selected. Please run 'rocketpool config' and try again.")
    }

    // Initialise command
    cmd := exec.Command("docker-compose", append([]string{"-f", filepath.Join(rpPath, "docker-compose.yml"), "--project-directory", rpPath}, args...)...)

    // Add environment variables
    env := []string{
        "COMPOSE_PROJECT_NAME=rocketpool",
        fmt.Sprintf("POW_CLIENT=%s",       rpConfig.GetSelectedEth1Client().Name),
        fmt.Sprintf("POW_IMAGE=%s",        rpConfig.GetSelectedEth1Client().Image),
        fmt.Sprintf("BEACON_CLIENT=%s",    rpConfig.GetSelectedEth2Client().Name),
        fmt.Sprintf("BEACON_IMAGE=%s",     rpConfig.GetSelectedEth2Client().Image),
        fmt.Sprintf("VALIDATOR_CLIENT=%s", rpConfig.GetSelectedEth2Client().Name),
        fmt.Sprintf("VALIDATOR_IMAGE=%s",  rpConfig.GetSelectedEth2Client().Image),
    }
    for _, param := range rpConfig.Chains.Eth1.Client.Params { env = append(env, param) }
    for _, param := range rpConfig.Chains.Eth2.Client.Params { env = append(env, param) }
    cmd.Env = append(os.Environ(), env...)

    // Return
    return cmd, nil

}


// Run a command and print its buffered output
func printOutput(cmd *exec.Cmd) error {

    // Get stdout & stderr pipes
    cmdOut, err := cmd.StdoutPipe()
    if err != nil { return err }
    cmdErr, err := cmd.StderrPipe()
    if err != nil { return err }

    // Print buffered stdout/stderr output
    go (func() {
        outScanner := bufio.NewScanner(cmdOut)
        for outScanner.Scan() { fmt.Println(outScanner.Text()) }
    })()
    go (func() {
        errScanner := bufio.NewScanner(cmdErr)
        for errScanner.Scan() { fmt.Println(errScanner.Text()) }
    })()

    // Run command
    if err := cmd.Run(); err != nil { return err }

    // Return
    return nil

}


// Run a command and return its output as lines
func readOutputLines(cmd *exec.Cmd) ([]string, error) {

    // Get stdout pipe
    cmdOut, err := cmd.StdoutPipe()
    if err != nil { return nil, err }

    // Start command
    if err := cmd.Start(); err != nil { return nil, err }

    // Read buffered output
    output := []string{}
    outScanner := bufio.NewScanner(cmdOut)
    for outScanner.Scan() { output = append(output, outScanner.Text()) }

    // Wait for command to complete
    if err := cmd.Wait(); err != nil { return nil, err }

    // Return
    return output, nil

}

