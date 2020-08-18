package rocketpool

import (
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "golang.org/x/crypto/ssh"

    "github.com/rocket-pool/smartnode/shared/services/config"
    "github.com/rocket-pool/smartnode/shared/utils/net"
)


// Config
const (
    RocketPoolPath = "~/.rocketpool"
    GlobalConfigFile = "config.yml"
    UserConfigFile = "settings.yml"
    ComposeFile = "docker-compose.yml"
)


// Rocket Pool client
type Client struct {
    client *ssh.Client
}


// Create new Rocket Pool client
func NewClient(hostAddress, user, keyPath string) (*Client, error) {

    // Initialize SSH client if configured for SSH
    var sshClient *ssh.Client
    if (hostAddress != "" && user != "" && keyPath != "") {

        // Read private key
        keyBytes, err := ioutil.ReadFile(keyPath)
        if err != nil {
            return nil, fmt.Errorf("Could not read SSH private key at %s: %w", keyPath, err)
        }

        // Parse private key
        key, err := ssh.ParsePrivateKey(keyBytes)
        if err != nil {
            return nil, fmt.Errorf("Could not parse SSH private key at %s: %w", keyPath, err)
        }

        // Initialise client
        sshClient, err = ssh.Dial("tcp", net.DefaultPort(hostAddress, "22"), &ssh.ClientConfig{
            User: user,
            Auth: []ssh.AuthMethod{ssh.PublicKeys(key)},
            HostKeyCallback: ssh.InsecureIgnoreHostKey(),
        })
        if err != nil {
            return nil, fmt.Errorf("Could not connect to %s as %s: %w", hostAddress, user, err)
        }

    }

    // Return client
    return &Client{
        client: sshClient,
    }, nil

}


// Close client remote connection
func (c *Client) Close() {
    if c.client != nil {
        c.client.Close()
    }
}


// Load the global config
func (c *Client) LoadGlobalConfig() (config.RocketPoolConfig, error) {
    return c.loadConfig(filepath.Join(RocketPoolPath, GlobalConfigFile))
}


// Save the user config
func (c *Client) SaveUserConfig(cfg config.RocketPoolConfig) error {
    return c.saveConfig(cfg, filepath.Join(RocketPoolPath, UserConfigFile))
}


// Start the Rocket Pool service
func (c *Client) StartService() error {
    cmd, err := c.compose("up -d")
    if err != nil { return err }
    return c.printOutput(cmd)
}


// Pause the Rocket Pool service
func (c *Client) PauseService() error {
    cmd, err := c.compose("stop")
    if err != nil { return err }
    return c.printOutput(cmd)
}


// Stop the Rocket Pool service
func (c *Client) StopService() error {
    cmd, err := c.compose("down -v")
    if err != nil { return err }
    return c.printOutput(cmd)
}


// Call the Rocket Pool API
func (c *Client) callAPI(args string) ([]byte, error) {
    cmd, err := c.compose(fmt.Sprintf("exec -T api /go/bin/rocketpool api %s", args))
    if err != nil { return []byte{}, err }
    return c.readOutput(cmd)
}


// Load a config file
func (c *Client) loadConfig(path string) (config.RocketPoolConfig, error) {

    // Read config
    configBytes, err := c.readOutput(fmt.Sprintf("cat %s", path))
    if err != nil {
        return config.RocketPoolConfig{}, fmt.Errorf("Could not read Rocket Pool config at %s: %w", path, err)
    }

    // Parse and return
    return config.Parse(configBytes)

}


// Save a config file
func (c *Client) saveConfig(cfg config.RocketPoolConfig, path string) error {

    // Serialize config
    configBytes, err := cfg.Serialize()
    if err != nil {
        return err
    }

    // Write config
    if _, err := c.readOutput(fmt.Sprintf("cat > %s <<EOF\n%sEOF", path, string(configBytes))); err != nil {
        return fmt.Errorf("Could not write Rocket Pool config to %s: %w", path, err)
    }

    // Return
    return nil

}


// Build a docker-compose command
func (c *Client) compose(args string) (string, error) {

    // Load config
    globalConfig, err := c.loadConfig(filepath.Join(RocketPoolPath, GlobalConfigFile))
    if err != nil {
        return "", err
    }
    userConfig, err := c.loadConfig(filepath.Join(RocketPoolPath, UserConfigFile))
    if err != nil {
        return "", err
    }
    rpConfig := config.Merge(&globalConfig, &userConfig)

    // Check config
    if rpConfig.GetSelectedEth1Client() == nil {
        return "", errors.New("No Eth 1.0 client selected. Please run 'rocketpool config' and try again.")
    }
    if rpConfig.GetSelectedEth2Client() == nil {
        return "", errors.New("No Eth 2.0 client selected. Please run 'rocketpool config' and try again.")
    }

    // Set environment variables from config
    env := []string{
        "COMPOSE_PROJECT_NAME=rocketpool",
        fmt.Sprintf("ETH1_CLIENT=%s",      rpConfig.GetSelectedEth1Client().ID),
        fmt.Sprintf("ETH1_IMAGE=%s",       rpConfig.GetSelectedEth1Client().Image),
        fmt.Sprintf("ETH2_CLIENT=%s",      rpConfig.GetSelectedEth2Client().ID),
        fmt.Sprintf("ETH2_IMAGE=%s",       rpConfig.GetSelectedEth2Client().Image),
        fmt.Sprintf("VALIDATOR_CLIENT=%s", rpConfig.GetSelectedEth2Client().ID),
        fmt.Sprintf("VALIDATOR_IMAGE=%s",  rpConfig.GetSelectedEth2Client().Image),
    }
    for _, param := range rpConfig.Chains.Eth1.Client.Params {
        env = append(env, fmt.Sprintf("%s=%s", param.Env, param.Value))
    }
    for _, param := range rpConfig.Chains.Eth2.Client.Params {
        env = append(env, fmt.Sprintf("%s=%s", param.Env, param.Value))
    }

    // Return command
    return fmt.Sprintf("%s docker-compose --project-directory %s -f %s %s", strings.Join(env, " "), RocketPoolPath, filepath.Join(RocketPoolPath, ComposeFile), args), nil

}


// Run a command and print its output
func (c *Client) printOutput(command string) error {
    if c.client == nil {

        // Initialize command
        cmd := exec.Command("sh", "-c", command)

        // Copy command output to stdout & stderr
        cmdOut, err := cmd.StdoutPipe()
        if err != nil { return err }
        cmdErr, err := cmd.StderrPipe()
        if err != nil { return err }
        go io.Copy(os.Stdout, cmdOut)
        go io.Copy(os.Stderr, cmdErr)

        // Run command
        return cmd.Run()

    } else {

        // Initialize session
        session, err := c.client.NewSession()
        if err != nil {
            return err
        }
        defer session.Close()

        // Copy session output to stdout & stderr
        cmdOut, err := session.StdoutPipe()
        if err != nil { return err }
        cmdErr, err := session.StderrPipe()
        if err != nil { return err }
        go io.Copy(os.Stdout, cmdOut)
        go io.Copy(os.Stderr, cmdErr)

        // Run command
        return session.Run(command)

    }
}


// Run a command and return its output
func (c *Client) readOutput(command string) ([]byte, error) {
    if c.client == nil {

        // Initialize command
        cmd := exec.Command("sh", "-c", command)

        // Run command and return output
        return cmd.Output()

    } else {

        // Initialize session
        session, err := c.client.NewSession()
        if err != nil {
            return []byte{}, err
        }
        defer session.Close()

        // Run command and return output
        return session.Output(command)

    }
}

