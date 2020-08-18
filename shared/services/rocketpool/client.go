package rocketpool

import (
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "os/exec"

    "golang.org/x/crypto/ssh"

    "github.com/rocket-pool/smartnode/shared/services/config"
    "github.com/rocket-pool/smartnode/shared/utils/net"
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

    // Read config
    configBytes, err := c.readOutput("cat ~/.rocketpool/config.yml")
    if err != nil {
        return config.RocketPoolConfig{}, fmt.Errorf("Could not read Rocket Pool config: %w", err)
    }

    // Parse and return
    return config.Parse(configBytes)

}


// Save the user config
func (c *Client) SaveUserConfig(cfg config.RocketPoolConfig) error {

    // Serialize config
    configBytes, err := cfg.Serialize()
    if err != nil {
        return err
    }

    // Write config
    if _, err := c.readOutput(fmt.Sprintf("cat > ~/.rocketpool/settings.yml <<EOF\n%s\nEOF", string(configBytes))); err != nil {
        return fmt.Errorf("Could not write Rocket Pool settings: %w", err)
    }

    // Return
    return nil

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

