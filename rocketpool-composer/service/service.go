package service

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Start Rocket Pool service
func startService() error {
    out, _ := compose("up", "-d").CombinedOutput()
    fmt.Println(string(out))
    return nil
}


// Pause Rocket Pool service
func pauseService() error {

    // Prompt for confirmation
    response := cliutils.Prompt(nil, nil, "Are you sure you want to pause the Rocket Pool services? Any staking minipools will be penalized! [y/n]", "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
    if strings.ToLower(response[:1]) == "n" { return nil }

    // Pause service
    out, _ := compose("stop").CombinedOutput()
    fmt.Println(string(out))
    return nil

}


// Stop Rocket Pool service
func stopService() error {

    // Prompt for confirmation
    response := cliutils.Prompt(nil, nil, "Are you sure you want to stop the Rocket Pool services? Any staking minipools will be penalized, and ethereum nodes will lose sync progress! [y/n]", "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
    if strings.ToLower(response[:1]) == "n" { return nil }

    // Stop service
    out, _ := compose("down", "-v").CombinedOutput()
    fmt.Println(string(out))
    return nil

}


// Scale Rocket Pool service
func scaleService(args ...string) error {
    out, _ := compose(append([]string{"scale"}, args...)...).CombinedOutput()
    fmt.Println(string(out))
    return nil
}


// Print Rocket Pool service logs
func serviceLogs(args ...string) error {
    out, _ := compose(append([]string{"logs"}, args...)...).CombinedOutput()
    fmt.Println(string(out))
    return nil
}


// Execute a Rocket Pool CLI command
func execCommand(args ...string) error {
    out, _ := compose(append([]string{"exec", "-T", "cli", "/go/bin/rocketpool", "run"}, args...)...).CombinedOutput()
    fmt.Println(string(out))
    return nil
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

