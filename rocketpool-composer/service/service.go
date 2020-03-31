package service

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)


// Start Rocket Pool service
func startService() error {
    out, _ := compose("up", "-d")
    fmt.Println(string(out))
    return nil
}


// Run a docker-compose subcommand with the given arguments and return combined output
func compose(args ...string) ([]byte, error) {

    // Initialise docker-compose command
    rpPath := os.Getenv("RP_PATH")
    cmd := exec.Command("docker-compose", append([]string{"-f", filepath.Join(rpPath, "docker-compose.yml"), "--project-directory", rpPath}, args...)...)

    // Add environment variables
    cmd.Env = append(os.Environ(),
        "COMPOSE_PROJECT_NAME=rocketpool",
        "POW_CLIENT=Infura",
        "POW_IMAGE=rocketpool/smartnode-pow-proxy:v0.0.1",
        "POW_INFURA_NETWORK=goerli",
        "POW_INFURA_PROJECT_ID=",
        "BEACON_CLIENT=Lighthouse",
        "BEACON_IMAGE=sigp/lighthouse:latest",
        "VALIDATOR_CLIENT=Lighthouse",
        "VALIDATOR_IMAGE=sigp/lighthouse:latest")

    // Run command & return output
    out, err := cmd.CombinedOutput()
    return out, err

}

