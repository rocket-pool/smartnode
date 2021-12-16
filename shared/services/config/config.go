package config

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"

	"github.com/imdario/mergo"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"

	"github.com/rocket-pool/rocketpool-go/utils/eth"
)

// Rocket Pool config
type RocketPoolConfig struct {
    Rocketpool struct {
        StorageAddress string           `yaml:"storageAddress,omitempty"`
        OneInchOracleAddress string     `yaml:"oneInchOracleAddress,omitempty"`
        RplTokenAddress string          `yaml:"rplTokenAddress,omitempty"`
        RPLFaucetAddress string         `yaml:"rplFaucetAddress,omitempty"`
    }                                   `yaml:"rocketpool,omitempty"`
    Smartnode struct {
        ProjectName string              `yaml:"projectName,omitempty"`
        GraffitiVersion string          `yaml:"graffitiVersion,omitempty"`
        Image string                    `yaml:"image,omitempty"`
        PasswordPath string             `yaml:"passwordPath,omitempty"`
        WalletPath string               `yaml:"walletPath,omitempty"`
        ValidatorKeychainPath string    `yaml:"validatorKeychainPath,omitempty"`
        ValidatorRestartCommand string  `yaml:"validatorRestartCommand,omitempty"`
        MaxFee float64                  `yaml:"maxFee,omitempty"`
        MaxPriorityFee float64          `yaml:"maxPriorityFee,omitempty"`
        GasLimit uint64                 `yaml:"gasLimit,omitempty"`
        RplClaimGasThreshold float64    `yaml:"rplClaimGasThreshold,omitempty"`
        TxWatchUrl string               `yaml:"txWatchUrl,omitempty"`
        StakeUrl string                 `yaml:"stakeUrl,omitempty"`
    }                                   `yaml:"smartnode,omitempty"`
    Chains struct {
        Eth1 Chain                      `yaml:"eth1,omitempty"`
        Eth1Fallback Chain              `yaml:"eth1Fallback,omitempty"`
        Eth2 Chain                      `yaml:"eth2,omitempty"`
    }                                   `yaml:"chains,omitempty"`
    Metrics Metrics                     `yaml:"metrics,omitempty"`
}
type Chain struct {
    Provider string                     `yaml:"provider,omitempty"`
    WsProvider string                   `yaml:"wsProvider,omitempty"`
    FallbackProvider string             `yaml:"fallbackProvider,omitempty"`
    FallbackWsProvider string           `yaml:"fallbackWsProvider,omitempty"`
    ReconnectDelay string               `yaml:"reconnectDelay,omitempty"`
    PruneProvisioner string             `yaml:"pruneProvisioner,omitempty"`
    ChainID string                      `yaml:"chainID,omitempty"`
    Client struct {
        Options []ClientOption          `yaml:"options,omitempty"`
        Selected string                 `yaml:"selected,omitempty"`
        Params []UserParam              `yaml:"params,omitempty"`
    }                                   `yaml:"client,omitempty"`
}
type ClientOption struct {
    ID string                           `yaml:"id,omitempty"`
    Name string                         `yaml:"name,omitempty"`
    Desc string                         `yaml:"desc,omitempty"`
    Image string                        `yaml:"image,omitempty"`
    BeaconImage string                  `yaml:"beaconImage,omitempty"`
    ValidatorImage string               `yaml:"validatorImage,omitempty"`
    Link string                         `yaml:"link,omitempty"`
    CompatibleEth2Clients string        `yaml:"compatibleEth2Clients,omitempty"`
    EventLogInterval string             `yaml:"eventLogInterval,omitempty"`
    Supermajority bool                  `yaml:"supermajority,omitempty"`
    Params []ClientParam                `yaml:"params,omitempty"`
    Fallback bool                       `yaml:"fallback,omitempty"`
}
type ClientParam struct {
    Name string                         `yaml:"name,omitempty"`
    Desc string                         `yaml:"desc,omitempty"`
    Env string                          `yaml:"env,omitempty"`
    Required bool                       `yaml:"required,omitempty"`
    Regex string                        `yaml:"regex,omitempty"`
    Type string                         `yaml:"type,omitempty"`
    Default string                      `yaml:"default,omitempty"`
    Max string                          `yaml:"max,omitempty"`
    BlankText string                    `yaml:"blankText,omitempty"`
    Advanced bool                       `yaml:"advanced,omitempty"`
}
type UserParam struct {
    Env string                          `yaml:"env,omitempty"`
    Value string                        `yaml:"value"`
}
type Metrics struct {
    Enabled bool                        `yaml:"enabled,omitempty"`
    Params []ClientParam                `yaml:"params,omitempty"`
    Settings []UserParam                `yaml:"settings,omitempty"`
}


// Get the selected clients from a config
func (config *RocketPoolConfig) GetSelectedEth1Client() *ClientOption {
    return config.Chains.Eth1.GetSelectedClient()
}
func (config *RocketPoolConfig) GetSelectedEth1FallbackClient() *ClientOption {
    return config.Chains.Eth1.GetClientById(config.Chains.Eth1Fallback.Client.Selected)
}
func (config *RocketPoolConfig) GetSelectedEth2Client() *ClientOption {
    return config.Chains.Eth2.GetSelectedClient()
}
func (chain *Chain) GetSelectedClient() *ClientOption {
    for _, option := range chain.Client.Options {
        if option.ID == chain.Client.Selected {
            return &option
        }
    }
    return nil
}

// Get a client by it's ID
func (chain *Chain) GetClientById(id string) *ClientOption {
    if id == "" {
        return nil
    }
    
    for _, option := range chain.Client.Options {
        if option.ID == id {
            return &option
        }
    }
    return nil
}


// Get a client parameter by its environment variable name
func (client *ClientOption) GetParamByEnvName(env string) *ClientParam {
    for _, param := range client.Params {
        if param.Env == env {
            return &param
        }
    }
    return nil
}


// Get a metrics parameter by its environment variable name
func (metrics *Metrics) GetParamByEnvName(env string) *ClientParam {
    for _, param := range metrics.Params {
        if param.Env == env {
            return &param
        }
    }
    return nil
}


// Get the beacon & validator images for a client
func (client *ClientOption) GetBeaconImage() string {
    if client.BeaconImage != "" {
        return client.BeaconImage
    } else {
        return client.Image
    }
}
func (client *ClientOption) GetValidatorImage() string {
    if client.ValidatorImage != "" {
        return client.ValidatorImage
    } else {
        return client.Image
    }
}


// Serialize a config to yaml bytes
func (config *RocketPoolConfig) Serialize() ([]byte, error) {
    bytes, err := yaml.Marshal(config)
    if err != nil {
        return []byte{}, fmt.Errorf("Could not serialize config: %w", err)
    }
    return bytes, nil
}


// Parse a config from yaml bytes
func Parse(bytes []byte) (RocketPoolConfig, error) {
    var config RocketPoolConfig
    if err := yaml.Unmarshal(bytes, &config); err != nil {
        return RocketPoolConfig{}, fmt.Errorf("Could not parse config: %w", err)
    }

    // Validate the defaults
    if err := ValidateDefaults(config.Chains.Eth1, "eth1"); err != nil {
        return RocketPoolConfig{}, err
    }
    if err := ValidateDefaults(config.Chains.Eth2, "eth2"); err != nil {
        return RocketPoolConfig{}, err
    }
    if err := ValidateMetricDefaults(config.Metrics.Params); err != nil {
        return RocketPoolConfig{}, err
    }

    return config, nil
}


// Make sure the default parameter values can be parsed into the parameter types
func ValidateDefaults(Chain Chain, ChainName string) error {
    for _, option := range Chain.Client.Options {
        for _, param := range option.Params {
            if param.Default != "" {
                var err error
                switch param.Type {
                    case "", "string":
                        continue
                    case "uint":
                        _, err = strconv.ParseUint(param.Default, 0, 0)
                    case "uint16":
                        _, err = strconv.ParseUint(param.Default, 0, 16)
                }
                if err != nil {
                    return fmt.Errorf("Could not parse config - " +
                        "parameter '%s' in %s client option '%s' " +
                        "is a %s but has a default value of '%s' which failed parsing: %w",
                        param.Name, ChainName, option.Name, param.Type, param.Default, err)
                }
            }
        }
    }
    return nil
}


// Make sure the default parameter values for the metrics section can be parsed into the parameter types
func ValidateMetricDefaults(Params []ClientParam) error {
    for _, param := range Params {
        if param.Default != "" {
            var err error
            switch param.Type {
                case "", "string":
                    continue
                case "uint":
                    _, err = strconv.ParseUint(param.Default, 0, 0)
                case "uint16":
                    _, err = strconv.ParseUint(param.Default, 0, 16)
            }
            if err != nil {
                return fmt.Errorf("Could not parse config - " +
                    "parameter '%s' in metrics " +
                    "is a %s but has a default value of '%s' which failed parsing: %w",
                    param.Name, param.Type, param.Default, err)
            }
        }
    }
    return nil
}


// Merge configs
func Merge(configs ...*RocketPoolConfig) (RocketPoolConfig, error) {
    var merged RocketPoolConfig
    for i := len(configs) - 1; i >= 0; i-- {
        if err := mergo.Merge(&merged, configs[i]); err != nil {
            return RocketPoolConfig{}, fmt.Errorf("Could not merge configs: %w", err)
        }
    }
    return merged, nil
}


// Load merged config from files
func Load(c *cli.Context) (RocketPoolConfig, error) {

    // Load configs
    globalConfig, err := loadFile(os.ExpandEnv(c.GlobalString("config")), true)
    if err != nil {
        return RocketPoolConfig{}, err
    }
    userConfig, err := loadFile(os.ExpandEnv(c.GlobalString("settings")), true)
    if err != nil {
        return RocketPoolConfig{}, err
    }
    cliConfig := getCliConfig(c)

    // Merge and return
    return Merge(&globalConfig, &userConfig, &cliConfig)

}


// Load config from a file
func loadFile(path string, required bool) (RocketPoolConfig, error) {

    // Read file; squelch not found errors if file is optional
    bytes, err := ioutil.ReadFile(path)
    if err != nil {
        if required {
            return RocketPoolConfig{}, fmt.Errorf("Could not find config file at %s: %w", path, err)
        } else {
            return RocketPoolConfig{}, nil
        }
    }

    // Parse config
    var config RocketPoolConfig
    if err := yaml.Unmarshal(bytes, &config); err != nil {
        return RocketPoolConfig{}, fmt.Errorf("Could not parse config file at %s: %w", path, err)
    }

    // Return
    return config, nil

}


// Create config from CLI arguments
func getCliConfig(c *cli.Context) RocketPoolConfig {
    var config RocketPoolConfig
    config.Rocketpool.StorageAddress = c.GlobalString("storageAddress")
    config.Rocketpool.OneInchOracleAddress = c.GlobalString("oneInchOracleAddress")
    config.Rocketpool.RplTokenAddress = c.GlobalString("rplTokenAddress")
    config.Rocketpool.RPLFaucetAddress = c.GlobalString("rplFaucetAddress")
    config.Smartnode.PasswordPath = c.GlobalString("password")
    config.Smartnode.WalletPath = c.GlobalString("wallet")
    config.Smartnode.ValidatorKeychainPath = c.GlobalString("validatorKeychain")
    config.Smartnode.MaxFee = c.GlobalFloat64("maxFee")
    config.Smartnode.MaxPriorityFee = c.GlobalFloat64("maxPrioFee")
    config.Smartnode.GasLimit = c.GlobalUint64("gasLimit")
    config.Chains.Eth1.Provider = c.GlobalString("eth1Provider")
    config.Chains.Eth2.Provider = c.GlobalString("eth2Provider")
    return config
}


// Parse and return the max fee in wei
func (config *RocketPoolConfig) GetMaxFee() (*big.Int, error) {

    // No gas price specified
    if config.Smartnode.MaxFee == 0 {
        return nil, nil
    }

    // Return gas price in wei
    return eth.GweiToWei(config.Smartnode.MaxFee), nil

}


// Parse and return the max priority fee in wei
func (config *RocketPoolConfig) GetMaxPriorityFee() (*big.Int, error) {

    // No gas price specified
    if config.Smartnode.MaxPriorityFee == 0 {
        return nil, nil
    }
    
    // Return gas price in wei
    return eth.GweiToWei(config.Smartnode.MaxPriorityFee), nil

}


// Parse and return the gas limit
func (config *RocketPoolConfig) GetGasLimit() (uint64, error) {

    // No gas limit specified
    if config.Smartnode.GasLimit == 0 {
        return 0, nil
    }

    // Return
    return config.Smartnode.GasLimit, nil

}

