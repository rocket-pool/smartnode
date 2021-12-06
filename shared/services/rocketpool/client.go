package rocketpool

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	osUser "os/user"
	"strings"
	"time"

	"github.com/a8m/envsubst"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
	kh "golang.org/x/crypto/ssh/knownhosts"

	"github.com/alessio/shellescape"
	"github.com/blang/semver/v4"
	externalip "github.com/glendc/go-external-ip"
	"github.com/mitchellh/go-homedir"
	"github.com/rocket-pool/smartnode/shared/services/config"
	"github.com/rocket-pool/smartnode/shared/utils/net"
)

// Config
const (
    InstallerURL = "https://github.com/rocket-pool/smartnode-install/releases/download/%s/install.sh"
    UpdateTrackerURL = "https://github.com/rocket-pool/smartnode-install/releases/download/%s/install-update-tracker.sh"

    GlobalConfigFile = "config.yml"
    UserConfigFile = "settings.yml"
    ComposeFile = "docker-compose.yml"
    MetricsComposeFile = "docker-compose-metrics.yml"
    PrometheusTemplate = "prometheus.tmpl"
    PrometheusFile = "prometheus.yml"

    APIContainerSuffix = "_api"
    APIBinPath = "/go/bin/rocketpool"

    DebugColor = color.FgYellow
)


// Rocket Pool client
type Client struct {
    configPath string
    daemonPath string
    maxFee float64
    maxPrioFee float64
    gasLimit uint64
    customNonce *big.Int
    client *ssh.Client
    originalMaxFee float64
    originalMaxPrioFee float64
    originalGasLimit uint64
    debugPrint bool
}


// Create new Rocket Pool client from CLI context
func NewClientFromCtx(c *cli.Context) (*Client, error) {
    return NewClient(c.GlobalString("config-path"), 
                     c.GlobalString("daemon-path"), 
                     c.GlobalString("host"), 
                     c.GlobalString("user"), 
                     c.GlobalString("key"), 
                     c.GlobalString("passphrase"),
                     c.GlobalString("known-hosts"),
                     c.GlobalFloat64("maxFee"),
                     c.GlobalFloat64("maxPrioFee"),
                     c.GlobalUint64("gasLimit"),
                     c.GlobalString("nonce"),
                     c.GlobalBool("debug"))
}


// Create new Rocket Pool client
func NewClient(configPath string, daemonPath string, hostAddress string, user string, keyPath string, passphrasePath string, knownhostsFile string, maxFee float64, maxPrioFee float64, gasLimit uint64, customNonce string, debug bool) (*Client, error) {

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
        keyBytes, err := ioutil.ReadFile(os.ExpandEnv(keyPath))
        if err != nil {
            return nil, fmt.Errorf("Could not read SSH private key at %s: %w", keyPath, err)
        }

        // Read passphrase
        var passphrase []byte
        if passphrasePath != "" {
            passphrase, err = ioutil.ReadFile(os.ExpandEnv(passphrasePath))
            if err != nil {
                return nil, fmt.Errorf("Could not read SSH passphrase at %s: %w", passphrasePath, err)
            }
        }

        // Parse private key
        var key ssh.Signer
        if passphrase == nil {
            key, err = ssh.ParsePrivateKey(keyBytes)
        } else {
            key, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, passphrase)
        }
        if err != nil {
            return nil, fmt.Errorf("Could not parse SSH private key at %s: %w", keyPath, err)
        }

        // Prepare the server host key callback function
        if knownhostsFile == "" {
            // Default to using the current users known_hosts file if one wasn't provided
            usr, err := osUser.Current()
            if err != nil {
                return nil, fmt.Errorf("Could not get current user: %w", err)
            }
            knownhostsFile = fmt.Sprintf("%s/.ssh/known_hosts", usr.HomeDir)
        }

        hostKeyCallback, err := kh.New(knownhostsFile)
        if err != nil {
            return nil, fmt.Errorf("Could not create hostKeyCallback function: %w", err)
        }

        // Initialise client
        sshClient, err = ssh.Dial("tcp", net.DefaultPort(hostAddress, "22"), &ssh.ClientConfig{
            User: user,
            Auth: []ssh.AuthMethod{ssh.PublicKeys(key)},
            HostKeyCallback: hostKeyCallback,
        })
        if err != nil {
            return nil, fmt.Errorf("Could not connect to %s as %s: %w", hostAddress, user, err)
        }

    }

    var customNonceBigInt *big.Int = nil
    var success bool
    if customNonce != "" {
        customNonceBigInt, success = big.NewInt(0).SetString(customNonce, 0)
        if !success {
            return nil, fmt.Errorf("Invalid nonce: %s", customNonce)
        }
    }

    // Return client
    return &Client{
        configPath: os.ExpandEnv(configPath),
        daemonPath: os.ExpandEnv(daemonPath),
        maxFee: maxFee,
        maxPrioFee: maxPrioFee,
        gasLimit: gasLimit,
        originalMaxFee: maxFee,
        originalMaxPrioFee: maxPrioFee,
        originalGasLimit: gasLimit,
        customNonce: customNonceBigInt,
        client: sshClient,
        debugPrint: debug,
    }, nil

}


// Close client remote connection
func (c *Client) Close() {
    if c.client != nil {
        _ = c.client.Close()
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

// Load the Prometheus template, do an environment variable substitution, and save it
func (c *Client) UpdatePrometheusConfiguration(settings []config.UserParam) error {
    prometheusTemplatePath, err := homedir.Expand(fmt.Sprintf("%s/%s", c.configPath, PrometheusTemplate))
    if err != nil {
        return fmt.Errorf("Error expanding Prometheus template path: %w", err)
    }
    
    prometheusConfigPath, err := homedir.Expand(fmt.Sprintf("%s/%s", c.configPath, PrometheusFile))
    if err != nil {
        return fmt.Errorf("Error expanding Prometheus config file path: %w", err)
    }

    // Set the environment variables defined in the user settings for metrics
    oldValues := map[string]string{}
    for _, setting := range settings {
        oldValues[setting.Env] = os.Getenv(setting.Env)
        os.Setenv(setting.Env, setting.Value)
    }

    // Read and substitute the template
    contents, err := envsubst.ReadFile(prometheusTemplatePath)
    if err != nil {
        return fmt.Errorf("Error reading and substituting Prometheus configuration template: %w", err)
    }

    // Unset the env vars
    for name, value := range oldValues {
        os.Setenv(name, value)
    }
    
    // Write the actual Prometheus config file
    err = ioutil.WriteFile(prometheusConfigPath, contents, 0664)
    if err != nil {
        return fmt.Errorf("Could not write Prometheus config file to %s: %w", shellescape.Quote(prometheusConfigPath), err)
    }
    err = os.Chmod(prometheusConfigPath, 0664)
    if err != nil {
        return fmt.Errorf("Could not set Prometheus config file permissions: %w", shellescape.Quote(prometheusConfigPath), err)
    }

    return nil
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
    return config.Merge(&globalConfig, &userConfig)
}


// Install the Rocket Pool service
func (c *Client) InstallService(verbose, noDeps bool, network, version, path string) error {

    // Get installation script downloader type
    downloader, err := c.getDownloader()
    if err != nil { return err }

    // Get installation script flags
    flags := []string{
        "-n", fmt.Sprintf("%s", shellescape.Quote(network)),
        "-v", fmt.Sprintf("%s", shellescape.Quote(version)),
    }
    if path != "" {
        flags = append(flags, fmt.Sprintf("-p %s", shellescape.Quote(path)))
    }
    if noDeps {
        flags = append(flags, "-d")
    }

    // Initialize installation command
    cmd, err := c.newCommand(fmt.Sprintf("%s %s | sh -s -- %s", downloader, fmt.Sprintf(InstallerURL, version), strings.Join(flags, " ")))
    if err != nil { return err }
    defer func() {
        _ = cmd.Close()
    }()

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
                _, _ = c.Println(scanner.Text())
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


// Install the update tracker
func (c *Client) InstallUpdateTracker(verbose bool, version string) error {

    // Get installation script downloader type
    downloader, err := c.getDownloader()
    if err != nil { return err }

    // Get installation script flags
    flags := []string{
        "-v", fmt.Sprintf("%s", shellescape.Quote(version)),
    }

    // Initialize installation command
    cmd, err := c.newCommand(fmt.Sprintf("%s %s | sh -s -- %s", downloader, fmt.Sprintf(UpdateTrackerURL, version), strings.Join(flags, " ")))
    if err != nil { return err }
    defer func() {
        _ = cmd.Close()
    }()

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
                _, _ = c.Println(scanner.Text())
            }
        }
    })()

    // Run command and return error output
    err = cmd.Run()
    if err != nil {
        return fmt.Errorf("Could not install Rocket Pool update tracker: %s", errMessage)
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
    sanitizedStrings := make([]string, len(serviceNames))
    for i, serviceName := range serviceNames {
        sanitizedStrings[i] = fmt.Sprintf("%s", shellescape.Quote(serviceName))
    }
    cmd, err := c.compose(composeFiles, fmt.Sprintf("logs -f --tail %s %s", shellescape.Quote(tail), strings.Join(sanitizedStrings, " ")))
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
        cmd = fmt.Sprintf("docker exec %s %s --version", shellescape.Quote(containerName), shellescape.Quote(APIBinPath))
    } else {
        cmd = fmt.Sprintf("%s --version", shellescape.Quote(c.daemonPath))
    }
    versionBytes, err := c.readOutput(cmd)
    if err != nil {
        return "", fmt.Errorf("Could not get Rocket Pool service version: %w", err)
    }

    // Get the version string
    outputString := string(versionBytes)
    elements := strings.Fields(outputString) // Split on whitespace
    if len(elements) < 1 {
        return "", fmt.Errorf("Could not parse Rocket Pool service version number from output '%s'", outputString)
    }
    versionString := elements[len(elements) - 1]

    // Make sure it's a semantic version
    version, err := semver.Make(versionString)
    if err != nil {
        return "", fmt.Errorf("Could not parse Rocket Pool service version number from output '%s': %w", outputString, err)
    }

    // Return the parsed semantic version (extra safety)
    return version.String(), nil

}


// Increments the custom nonce parameter.
// This is used for calls that involve multiple transactions, so they don't all have the same nonce.
func (c *Client) IncrementCustomNonce() {
    c.customNonce.Add(c.customNonce, big.NewInt(1))
}


// Get the current Docker image used by the given container
func (c *Client) GetDockerImage(container string) (string, error) {

    cmd := fmt.Sprintf("docker container inspect --format={{.Config.Image}} %s", container)
    image, err := c.readOutput(cmd)
    if err != nil {
        return "", err
    }

    return strings.TrimSpace(string(image)), nil

}


// Get the current Docker image used by the given container
func (c *Client) GetDockerStatus(container string) (string, error) {

    cmd := fmt.Sprintf("docker container inspect --format={{.State.Status}} %s", container)
    status, err := c.readOutput(cmd)
    if err != nil {
        return "", err
    }

    return strings.TrimSpace(string(status)), nil

}


// Get the time that the given container shut down
func (c *Client) GetDockerContainerShutdownTime(container string) (time.Time, error) {

    cmd := fmt.Sprintf("docker container inspect --format={{.State.FinishedAt}} %s", container)
    finishTimeBytes, err := c.readOutput(cmd)
    if err != nil {
        return time.Time{}, err
    }

    finishTime, err := time.Parse(time.RFC3339, strings.TrimSpace(string(finishTimeBytes)))
    if err != nil {
        return time.Time{}, fmt.Errorf("Error parsing validator container exit time [%s]: %w", string(finishTimeBytes), err)
    }

    return finishTime, nil
    
}


// Shut down a container
func (c *Client) StopContainer(container string) (string, error) {

    cmd := fmt.Sprintf("docker stop %s", container)
    output, err := c.readOutput(cmd)
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
    
}


// Get the gas settings
func (c *Client) GetGasSettings() (float64, float64, uint64) {
    return c.maxFee, c.maxPrioFee, c.gasLimit
}


// Get the gas fees
func (c *Client) AssignGasSettings(maxFee float64, maxPrioFee float64, gasLimit uint64) {
    c.maxFee = maxFee
    c.maxPrioFee = maxPrioFee
    c.gasLimit = gasLimit
}


// Load a config file
func (c *Client) loadConfig(path string) (config.RocketPoolConfig, error) {
    expandedPath, err := homedir.Expand(path)
    if err != nil {
        return config.RocketPoolConfig{}, err
    }
    configBytes, err := ioutil.ReadFile(expandedPath)
    if err != nil {
        return config.RocketPoolConfig{}, fmt.Errorf("Could not read Rocket Pool config at %s: %w", shellescape.Quote(path), err)
    }
    return config.Parse(configBytes)
}


// Save a config file
func (c *Client) saveConfig(cfg config.RocketPoolConfig, path string) error {
    configBytes, err := cfg.Serialize()
    if err != nil {
        return err
    }
    expandedPath, err := homedir.Expand(path)
    if err != nil {
        return err
    }
    if err := ioutil.WriteFile(expandedPath, configBytes, 0); err != nil {
        return fmt.Errorf("Could not write Rocket Pool config to %s: %w", shellescape.Quote(expandedPath), err)
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
    eth1Client := cfg.GetSelectedEth1Client()
    eth2Client := cfg.GetSelectedEth2Client()
    if eth1Client == nil {
        return "", errors.New("No Eth 1.0 client selected. Please run 'rocketpool service config' and try again.")
    }
    if eth2Client == nil {
        return "", errors.New("No Eth 2.0 client selected. Please run 'rocketpool service config' and try again.")
    }

    // Make sure the selected eth2 is compatible with the selected eth1
    isCompatible := false 
    if eth1Client.CompatibleEth2Clients == "" {
        isCompatible = true
    } else {
        compatibleEth2ClientIds := strings.Split(eth1Client.CompatibleEth2Clients, ";")
        for _, id := range compatibleEth2ClientIds {
            if id == eth2Client.ID {
                isCompatible = true
                break
            }
        }
    }
    if !isCompatible {
        return "", fmt.Errorf("Eth 2.0 client [%s] is incompatible with Eth 1.0 client [%s]. Please run 'rocketpool service config' and select compatible clients.", eth2Client.Name, eth1Client.Name)
    }

    // Get the external IP address
    var externalIP string
    consensus := externalip.DefaultConsensus(nil, nil)
    ip, err := consensus.ExternalIP()
    if err != nil {
        fmt.Println("Warning: couldn't get external IP address; if you're using Nimbus, it may have trouble finding peers:")
        fmt.Println(err.Error())
    } else {
        externalIP = ip.String()
    }

    // Set environment variables from config
    env := []string{
        fmt.Sprintf("COMPOSE_PROJECT_NAME=%s",    shellescape.Quote(cfg.Smartnode.ProjectName)),
        fmt.Sprintf("ROCKET_POOL_VERSION=%s",     shellescape.Quote(cfg.Smartnode.GraffitiVersion)),
        fmt.Sprintf("SMARTNODE_IMAGE=%s",         shellescape.Quote(cfg.Smartnode.Image)),
        fmt.Sprintf("ETH1_CLIENT=%s",             shellescape.Quote(cfg.GetSelectedEth1Client().ID)),
        fmt.Sprintf("ETH1_IMAGE=%s",              shellescape.Quote(cfg.GetSelectedEth1Client().Image)),
        fmt.Sprintf("ETH2_CLIENT=%s",             shellescape.Quote(cfg.GetSelectedEth2Client().ID)),
        fmt.Sprintf("ETH2_IMAGE=%s",              shellescape.Quote(cfg.GetSelectedEth2Client().GetBeaconImage())),
        fmt.Sprintf("VALIDATOR_CLIENT=%s",        shellescape.Quote(cfg.GetSelectedEth2Client().ID)),
        fmt.Sprintf("VALIDATOR_IMAGE=%s",         shellescape.Quote(cfg.GetSelectedEth2Client().GetValidatorImage())),
        fmt.Sprintf("ETH1_PROVIDER=%s",           shellescape.Quote(cfg.Chains.Eth1.Provider)),
        fmt.Sprintf("ETH1_WS_PROVIDER=%s",        shellescape.Quote(cfg.Chains.Eth1.WsProvider)),
        fmt.Sprintf("ETH2_PROVIDER=%s",           shellescape.Quote(cfg.Chains.Eth2.Provider)),
        fmt.Sprintf("EXTERNAL_IP=%s",             shellescape.Quote(externalIP)),
    }
    if cfg.Metrics.Enabled {
        env = append(env, "ENABLE_METRICS=1")
    } else {
        env = append(env, "ENABLE_METRICS=0")
    }
    paramsSet := map[string]bool{}
    for _, param := range cfg.Chains.Eth1.Client.Params {
        env = append(env, fmt.Sprintf("%s=%s", param.Env, shellescape.Quote(param.Value)))
        paramsSet[param.Env] = true
    }
    for _, param := range cfg.Chains.Eth2.Client.Params {
        env = append(env, fmt.Sprintf("%s=%s", param.Env, shellescape.Quote(param.Value)))
        paramsSet[param.Env] = true
    }
    for _, setting := range cfg.Metrics.Settings {
        env = append(env, fmt.Sprintf("%s=%s", setting.Env, shellescape.Quote(setting.Value)))
        paramsSet[setting.Env] = true
    }

    // Set default values from client config
    for _, param := range cfg.GetSelectedEth1Client().Params {
        if _, ok := paramsSet[param.Env]; ok { continue }
        if param.Default == "" { continue }
        env = append(env, fmt.Sprintf("%s=%s", param.Env, shellescape.Quote(param.Default)))
    }
    for _, param := range cfg.GetSelectedEth2Client().Params {
        if _, ok := paramsSet[param.Env]; ok { continue }
        if param.Default == "" { continue }
        env = append(env, fmt.Sprintf("%s=%s", param.Env, shellescape.Quote(param.Default)))
    }
    for _, param := range cfg.Metrics.Params {
        if _, ok := paramsSet[param.Env]; ok { continue }
        if param.Default == "" { continue }
        env = append(env, fmt.Sprintf("%s=%s", param.Env, shellescape.Quote(param.Default)))
    }

    // How many built-in compose files are we using
    builtInFileCount := 1
    if cfg.Metrics.Enabled {
        builtInFileCount = 2
    }

    // Set compose file flags
    composeFileFlags := make([]string, len(composeFiles) + builtInFileCount)
    expandedConfigPath, err := homedir.Expand(c.configPath)
    if err != nil {
        return "", err
    }

    // Add the default docker-compose.yml
    composeFileFlags[0] = fmt.Sprintf("-f %s", shellescape.Quote((fmt.Sprintf("%s/%s", expandedConfigPath, ComposeFile))))

    // Add docker-compose-metrics.yml if metrics are enabled
    if cfg.Metrics.Enabled {
        composeFileFlags[1] = fmt.Sprintf("-f %s", shellescape.Quote((fmt.Sprintf("%s/%s", expandedConfigPath, MetricsComposeFile))))
    }

    for fi, composeFile := range composeFiles {
        expandedFile, err := homedir.Expand(composeFile)
        if err != nil {
            return "", err
        }
        composeFileFlags[fi+builtInFileCount] = fmt.Sprintf("-f %s", shellescape.Quote(expandedFile))
    }

    // Return command
    return fmt.Sprintf("%s docker-compose --project-directory %s %s %s", strings.Join(env, " "), shellescape.Quote(expandedConfigPath), strings.Join(composeFileFlags, " "), args), nil

}


// Call the Rocket Pool API
func (c *Client) callAPI(args string, otherArgs ...string) ([]byte, error) {
    // Sanitize arguments
    var sanitizedArgs []string
    for _, arg := range strings.Fields(args) {
        sanitizedArg := shellescape.Quote(arg)
        sanitizedArgs = append(sanitizedArgs, sanitizedArg)
    }
    args = strings.Join(sanitizedArgs, " ")
    if len(otherArgs) > 0 {
        for _, arg := range otherArgs {
            sanitizedArg := shellescape.Quote(arg)
            args += fmt.Sprintf(" %s", sanitizedArg)
        }
    }

    // Run the command
    var cmd string
    if c.daemonPath == "" {
        containerName, err := c.getAPIContainerName()
        if err != nil {
            return []byte{}, err
        }
        cmd = fmt.Sprintf("docker exec %s %s %s %s api %s", shellescape.Quote(containerName), shellescape.Quote(APIBinPath), c.getGasOpts(), c.getCustomNonce(), args)
    } else {
        cmd = fmt.Sprintf("%s --config %s --settings %s %s %s api %s", 
            c.daemonPath, 
            shellescape.Quote(fmt.Sprintf("%s/%s", c.configPath, GlobalConfigFile)), 
            shellescape.Quote(fmt.Sprintf("%s/%s", c.configPath, UserConfigFile)),
            c.getGasOpts(),
            c.getCustomNonce(),
            args)
    }
    
    if c.debugPrint {
        fmt.Println("To API:")
        fmt.Println(cmd)
    }

    output, err := c.readOutput(cmd)
    
    if c.debugPrint {
        if output != nil {
            fmt.Println("API Out:")
            fmt.Println(string(output))
        }
        if err != nil {
            fmt.Println("API Err:")
            fmt.Println(err.Error())
        }
    }

    // Reset the gas settings after the call
    c.maxFee = c.originalMaxFee
    c.maxPrioFee = c.originalMaxPrioFee
    c.gasLimit = c.originalGasLimit

    return output, err
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


// Get gas price & limit flags
func (c *Client) getGasOpts() string {
    var opts string
    opts += fmt.Sprintf("--maxFee %f ", c.maxFee)
    opts += fmt.Sprintf("--maxPrioFee %f ", c.maxPrioFee)
    opts += fmt.Sprintf("--gasLimit %d ", c.gasLimit)
    return opts
}


func (c *Client) getCustomNonce() string {
    // Set the custom nonce
    nonce := ""
    if c.customNonce != nil {
        nonce = fmt.Sprintf("--nonce %s", c.customNonce.String())
    }
    return nonce
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
    defer func() {
        _ = cmd.Close()
    }()

    // Copy command output to stdout & stderr
    cmdOut, err := cmd.StdoutPipe()
    if err != nil { return err }
    cmdErr, err := cmd.StderrPipe()
    if err != nil { return err }
    go func() {
        _, err := io.Copy(os.Stdout, cmdOut)
        if err != nil {
        	log.Printf("Error piping stdout: %v", err)
        }
    }()
    go func() {
        _, err := io.Copy(os.Stderr, cmdErr)
        if err != nil {
            log.Printf("Error piping stderr: %v", err)
        }
    }()

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
    defer func() {
        _ = cmd.Close()
    }()

    // Run command and return output
    return cmd.Output()

}

