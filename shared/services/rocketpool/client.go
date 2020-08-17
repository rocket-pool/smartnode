package rocketpool

import (
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "os/exec"
    "strings"

    "golang.org/x/crypto/ssh"

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


// Run a command and print its output
func (c *Client) printOutput(command string, args, env []string) error {
    if c.client == nil {

        // Initialize command
        cmd := exec.Command(command, args...)

        // Copy command output to stdout & stderr
        cmdOut, err := cmd.StdoutPipe()
        if err != nil { return err }
        cmdErr, err := cmd.StderrPipe()
        if err != nil { return err }
        go io.Copy(os.Stdout, cmdOut)
        go io.Copy(os.Stderr, cmdErr)

        // Run command
        if err := cmd.Run(); err != nil {
            return err
        }

    } else {

        // Initialize session
        sess, err := c.client.NewSession()
        if err != nil {
            return err
        }
        defer sess.Close()

        // Copy session output to stdout & stderr
        cmdOut, err := sess.StdoutPipe()
        if err != nil { return err }
        cmdErr, err := sess.StderrPipe()
        if err != nil { return err }
        go io.Copy(os.Stdout, cmdOut)
        go io.Copy(os.Stderr, cmdErr)

        // Run command
        if err := sess.Run(fmt.Sprintf("%s %s", command, strings.Join(args, " "))); err != nil {
            return err
        }

    }
    return nil
}


// Run a command and return its output
func (c *Client) readOutput(command string, args, env []string) ([]byte, error) {
    if c.client == nil {

        // Initialize command
        cmd := exec.Command(command, args...)

        // Run command and return output
        return cmd.Output()

    } else {

        // Initialize session
        sess, err := c.client.NewSession()
        if err != nil {
            return []byte{}, err
        }
        defer sess.Close()

        // Run command and return output
        return sess.Output(fmt.Sprintf("%s %s", command, strings.Join(args, " ")))

    }
}

