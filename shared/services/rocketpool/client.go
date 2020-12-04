package rocketpool

import (
    "bufio"
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "regexp"
    "strings"

    "github.com/fatih/color"
    "github.com/urfave/cli"
    "golang.org/x/crypto/ssh"

    "github.com/rocket-pool/smartnode/shared/services/config"
    "github.com/rocket-pool/smartnode/shared/utils/net"
)


// Config
const (
    InstallerURL = "https://github.com/rocket-pool/smartnode-install/releases/latest/download/install.sh"

    GlobalConfigFile = "config.yml"
    UserConfigFile = "settings.yml"
    ComposeFile = "docker-compose.yml"

    APIContainerSuffix = "_api"
    APIBinPath = "/go/bin/rocketpool"

    DebugColor = color.FgYellow
)


// Rocket Pool client
type Client struct {
    configPath string
    daemonPath string
    client *ssh.Client
}


// Create new Rocket Pool client from CLI context
func NewClientFromCtx(c *cli.Context) (*Client, error) {
    return NewClient(c.GlobalString("config-path"), c.GlobalString("daemon-path"), c.GlobalString("host"), c.GlobalString("user"), c.GlobalString("key"), c.GlobalString("passphrase"))
}


// Create new Rocket Pool client
func NewClient(configPath, daemonPath, hostAddress, user, keyPath, keyPassphrase string) (*Client, error) {

    // Initialize SSH client if configured for SSH
    var sshClient *ssh.Client
    if hostAddress != "" {

        // Check parameters
        if user == "" {
            return nil, errors.New("The SSH user (--user) must be specified.")
        }
        if keyPath == "" {
            return nil, errors.New("The SSH private key path (--key) must be specified.")
        }

        // Read private key
        keyBytes, err := ioutil.ReadFile(keyPath)
        if err != nil {
            return nil, fmt.Errorf("Could not read SSH private key at %s: %w", keyPath, err)
        }

        // Parse private key
        var key ssh.Signer
        if keyPassphrase == "" {
            key, err = ssh.ParsePrivateKey(keyBytes)
        } else {
            key, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(keyPassphrase))
        }
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
        configPath: configPath,
        daemonPath: daemonPath,
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
    return c.loadConfig(fmt.Sprintf("%s/%s", c.configPath, GlobalConfigFile))
}


// Load/save the user config
func (c *Client) LoadUserConfig() (config.RocketPoolConfig, error) {
    return c.loadConfig(fmt.Sprintf("%s/%s", c.configPath, UserConfigFile))
}
func (c *Client) SaveUserConfig(cfg config.RocketPoolConfig) error {
    return c.saveConfig(cfg, fmt.Sprintf("%s/%s", c.configPath, UserConfigFile))
}


// Load the merged global & user config
func (c *Client) LoadMergedConfig() (config.RocketPoolConfig, error) {
    globalConfig, err := c.LoadGlobalConfig()
    if err != nil {
        return config.RocketPoolConfig{}, err
    }
    userConfig, err := c.LoadUserConfig()
    if err != nil {
        return config.RocketPoolConfig{}, err
    }
    return config.Merge(&globalConfig, &userConfig), nil
}


// Install the Rocket Pool service
func (c *Client) InstallService(verbose, noDeps bool, network, version string) error {

    // Get installation script downloader type
    downloader, err := c.getDownloader()
    if err != nil { return err }

    // Get installation script flags
    flags := []string{
        "-n", network,
        "-v", version,
    }
    if noDeps {
        flags = append(flags, "-d")
    }

    // Initialize installation command
    cmd, err := c.newCommand(fmt.Sprintf("%s %s | sh -s -- %s", downloader, InstallerURL, strings.Join(flags, " ")))
    if err != nil { return err }
    defer cmd.Close()

    // Get command output pipes
    cmdOut, err := cmd.StdoutPipe()
    if err != nil { return err }
    cmdErr, err := cmd.StderrPipe()
    if err != nil { return err }

    // Print progress from stdout
    go (func() {
        scanner := bufio.NewScanner(cmdOut)
        for scanner.Scan() {
            fmt.Println(scanner.Text())
        }
    })()

    // Read command & error output from stderr; render in verbose mode
    var errMessage string
    go (func() {
        c := color.New(DebugColor)
        scanner := bufio.NewScanner(cmdErr)
        for scanner.Scan() {
            errMessage = scanner.Text()
            if verbose {
                c.Println(scanner.Text())
            }
        }
    })()

    // Run command and return error output
    err = cmd.Run()
    if err != nil {
        return fmt.Errorf("Could not install Rocket Pool service: %s", errMessage)
    }
    return nil

}


// Start the Rocket Pool service
func (c *Client) StartService(composeFiles []string) error {
    cmd, err := c.compose(composeFiles, "up -d")
    if err != nil { return err }
    return c.printOutput(cmd)
}


// Pause the Rocket Pool service
func (c *Client) PauseService(composeFiles []string) error {
    cmd, err := c.compose(composeFiles, "stop")
    if err != nil { return err }
    return c.printOutput(cmd)
}


// Stop the Rocket Pool service
func (c *Client) StopService(composeFiles []string) error {
    cmd, err := c.compose(composeFiles, "down -v")
    if err != nil { return err }
    return c.printOutput(cmd)
}


// Print the Rocket Pool service status
func (c *Client) PrintServiceStatus(composeFiles []string) error {
    cmd, err := c.compose(composeFiles, "ps")
    if err != nil { return err }
    return c.printOutput(cmd)
}


// Print the Rocket Pool service logs
func (c *Client) PrintServiceLogs(composeFiles []string, tail string, serviceNames ...string) error {
    cmd, err := c.compose(composeFiles, fmt.Sprintf("logs -f --tail %s %s", tail, strings.Join(serviceNames, " ")))
    if err != nil { return err }
    return c.printOutput(cmd)
}


// Print the Rocket Pool service stats
func (c *Client) PrintServiceStats(composeFiles []string) error {

    // Get service container IDs
    cmd, err := c.compose(composeFiles, "ps -q")
    if err != nil { return err }
    containers, err := c.readOutput(cmd)
    if err != nil { return err }
    containerIds := strings.Split(strings.TrimSpace(string(containers)), "\n")

    // Print stats
    return c.printOutput(fmt.Sprintf("docker stats %s", strings.Join(containerIds, " ")))

}


// Get the Rocket Pool service version
func (c *Client) GetServiceVersion() (string, error) {

    // Get service container version output
    var cmd string
    if c.daemonPath == "" {
        containerName, err := c.getAPIContainerName()
        if err != nil {
            return "", err
        }
        cmd = fmt.Sprintf("docker exec %s %s --version", containerName, APIBinPath)
    } else {
        cmd = fmt.Sprintf("%s --version", c.daemonPath)
    }
    versionBytes, err := c.readOutput(cmd)
    if err != nil {
        return "", fmt.Errorf("Could not get Rocket Pool service version: %w", err)
    }

    // Parse version number
    versionNumberBytes := regexp.MustCompile("v?(\\d+\\.)*\\d+").Find(versionBytes)
    if versionNumberBytes == nil {
        return "", errors.New("Could not parse Rocket Pool service version number.")
    }

    // Return
    return string(versionNumberBytes), nil

}


// Load a config file
func (c *Client) loadConfig(path string) (config.RocketPoolConfig, error) {
    configBytes, err := c.readOutput(fmt.Sprintf("cat %s", path))
    if err != nil {
        return config.RocketPoolConfig{}, fmt.Errorf("Could not read Rocket Pool config at %s: %w", path, err)
    }
    return config.Parse(configBytes)
}


// Save a config file
func (c *Client) saveConfig(cfg config.RocketPoolConfig, path string) error {
    configBytes, err := cfg.Serialize()
    if err != nil {
        return err
    }
    if _, err := c.readOutput(fmt.Sprintf("cat > %s <<EOF\n%sEOF", path, string(configBytes))); err != nil {
        return fmt.Errorf("Could not write Rocket Pool config to %s: %w", path, err)
    }
    return nil
}


// Build a docker-compose command
func (c *Client) compose(composeFiles []string, args string) (string, error) {

    // Cancel if running in non-docker mode
    if c.daemonPath != "" {
        return "", errors.New("Command unavailable with '--daemon-path' option specified.")
    }

    // Load config
    cfg, err := c.LoadMergedConfig()
    if err != nil {
        return "", err
    }

    // Check config
    if cfg.GetSelectedEth1Client() == nil {
        return "", errors.New("No Eth 1.0 client selected. Please run 'rocketpool service config' and try again.")
    }
    if cfg.GetSelectedEth2Client() == nil {
        return "", errors.New("No Eth 2.0 client selected. Please run 'rocketpool service config' and try again.")
    }

    // Set environment variables from config
    env := []string{
        fmt.Sprintf("COMPOSE_PROJECT_NAME='%s'",    cfg.Smartnode.ProjectName),
        fmt.Sprintf("SMARTNODE_IMAGE='%s'",         cfg.Smartnode.Image),
        fmt.Sprintf("ETH1_CLIENT='%s'",             cfg.GetSelectedEth1Client().ID),
        fmt.Sprintf("ETH1_IMAGE='%s'",              cfg.GetSelectedEth1Client().Image),
        fmt.Sprintf("ETH2_CLIENT='%s'",             cfg.GetSelectedEth2Client().ID),
        fmt.Sprintf("ETH2_IMAGE='%s'",              cfg.GetSelectedEth2Client().GetBeaconImage()),
        fmt.Sprintf("VALIDATOR_CLIENT='%s'",        cfg.GetSelectedEth2Client().ID),
        fmt.Sprintf("VALIDATOR_IMAGE='%s'",         cfg.GetSelectedEth2Client().GetValidatorImage()),
        fmt.Sprintf("ETH1_PROVIDER='%s'",           cfg.Chains.Eth1.Provider),
        fmt.Sprintf("ETH2_PROVIDER='%s'",           cfg.Chains.Eth2.Provider),
    }
    for _, param := range cfg.Chains.Eth1.Client.Params {
        env = append(env, fmt.Sprintf("%s='%s'", param.Env, param.Value))
    }
    for _, param := range cfg.Chains.Eth2.Client.Params {
        env = append(env, fmt.Sprintf("%s='%s'", param.Env, param.Value))
    }

    // Set compose file flags
    composeFileFlags := make([]string, len(composeFiles) + 1)
    composeFileFlags[0] = fmt.Sprintf("-f %s/%s", c.configPath, ComposeFile)
    for fi, composeFile := range composeFiles {
        composeFileFlags[fi + 1] = fmt.Sprintf("-f %s", composeFile)
    }

    // Return command
    return fmt.Sprintf("%s docker-compose --project-directory %s %s %s", strings.Join(env, " "), c.configPath, strings.Join(composeFileFlags, " "), args), nil

}


// Call the Rocket Pool API
func (c *Client) callAPI(args string) ([]byte, error) {
    var cmd string
    if c.daemonPath == "" {
        containerName, err := c.getAPIContainerName()
        if err != nil {
            return []byte{}, err
        }
        cmd = fmt.Sprintf("docker exec %s %s api %s", containerName, APIBinPath, args)
    } else {
        cmd = fmt.Sprintf("%s --config %s --settings %s api %s", c.daemonPath, fmt.Sprintf("%s/%s", c.configPath, GlobalConfigFile), fmt.Sprintf("%s/%s", c.configPath, UserConfigFile), args)
    }
    return c.readOutput(cmd)
}


// Get the API container name
func (c *Client) getAPIContainerName() (string, error) {
    cfg, err := c.LoadMergedConfig()
    if err != nil {
        return "", err
    }
    if cfg.Smartnode.ProjectName == "" {
      return "", errors.New("Rocket Pool docker project name not set")
    }
    return cfg.Smartnode.ProjectName + APIContainerSuffix, nil
}


// Get the first downloader available to the system
func (c *Client) getDownloader() (string, error) {

    // Check for cURL
    hasCurl, err := c.readOutput("command -v curl")
    if err == nil && len(hasCurl) > 0 {
        return "curl -sL", nil
    }

    // Check for wget
    hasWget, err := c.readOutput("command -v wget")
    if err == nil && len(hasWget) > 0 {
        return "wget -qO-", nil
    }

    // Return error
    return "", errors.New("Either cURL or wget is required to begin installation.")

}


// Run a command and print its output
func (c *Client) printOutput(cmdText string) error {

    // Initialize command
    cmd, err := c.newCommand(cmdText)
    if err != nil { return err }
    defer cmd.Close()

    // Copy command output to stdout & stderr
    cmdOut, err := cmd.StdoutPipe()
    if err != nil { return err }
    cmdErr, err := cmd.StderrPipe()
    if err != nil { return err }
    go io.Copy(os.Stdout, cmdOut)
    go io.Copy(os.Stderr, cmdErr)

    // Run command
    return cmd.Run()

}


// Run a command and return its output
func (c *Client) readOutput(cmdText string) ([]byte, error) {

    // Initialize command
    cmd, err := c.newCommand(cmdText)
    if err != nil {
        return []byte{}, err
    }
    defer cmd.Close()

    // Run command and return output
    return cmd.Output()

}

