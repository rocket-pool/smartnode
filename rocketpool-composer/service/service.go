package service

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Start Rocket Pool service
func startService() error {
    return printOutput(compose("up", "-d"))
}


// Pause Rocket Pool service
func pauseService() error {

    // Prompt for confirmation
    response := cliutils.Prompt(nil, nil, "Are you sure you want to pause the Rocket Pool services? Any staking minipools will be penalized! [y/n]", "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
    if strings.ToLower(response[:1]) == "n" { return nil }

    // Pause service
    return printOutput(compose("stop"))

}


// Stop Rocket Pool service
func stopService() error {

    // Prompt for confirmation
    response := cliutils.Prompt(nil, nil, "Are you sure you want to stop the Rocket Pool services? Any staking minipools will be penalized, and ethereum nodes will lose sync progress! [y/n]", "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
    if strings.ToLower(response[:1]) == "n" { return nil }

    // Stop service
    return printOutput(compose("down", "-v"))

}


// Scale Rocket Pool service
func scaleService(args ...string) error {
    return printOutput(compose(append([]string{"scale"}, args...)...))
}


// Print Rocket Pool service logs
func serviceLogs(args ...string) error {
    return printOutput(compose(append([]string{"logs"}, args...)...))
}


// Execute a Rocket Pool CLI command
func execCommand(args ...string) error {
    return printOutput(compose(append([]string{"exec", "-T", "cli", "/go/bin/rocketpool", "run"}, args...)...))
}


// Build a docker-compose command with the given arguments
func compose(args ...string) *exec.Cmd {

    // Initialise command
    rpPath := os.Getenv("RP_PATH")
    cmd := exec.Command("docker-compose", append([]string{"-f", filepath.Join(rpPath, "docker-compose.yml"), "--project-directory", rpPath}, args...)...)

    // Add environment variables
    cmd.Env = append(os.Environ(),
        "COMPOSE_PROJECT_NAME=rocketpool",
        "POW_CLIENT=Infura",
        "POW_IMAGE=rocketpool/smartnode-pow-proxy:v0.0.1",
        "POW_INFURA_NETWORK=goerli",
        "POW_INFURA_PROJECT_ID=d690a0156a994dd785c0a64423586f52",
        "BEACON_CLIENT=Lighthouse",
        "BEACON_IMAGE=sigp/lighthouse:latest",
        "VALIDATOR_CLIENT=Lighthouse",
        "VALIDATOR_IMAGE=sigp/lighthouse:latest")

    // Return
    return cmd

}


// Run a command and print its buffered stdout/stderr output
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

