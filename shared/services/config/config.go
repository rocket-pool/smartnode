package config

import (
    "fmt"
    "io/ioutil"
    "math/big"
    "os"

    "github.com/imdario/mergo"
    "github.com/urfave/cli"
    "gopkg.in/yaml.v2"

    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Rocket Pool config
type RocketPoolConfig struct {
    Rocketpool struct {
        StorageAddress string           `yaml:"storageAddress,omitempty"`
    }                                   `yaml:"rocketpool,omitempty"`
    Smartnode struct {
        ProjectName string              `yaml:"projectName,omitempty"`
        Image string                    `yaml:"image,omitempty"`
        PasswordPath string             `yaml:"passwordPath,omitempty"`
        WalletPath string               `yaml:"walletPath,omitempty"`
        ValidatorKeychainPath string    `yaml:"validatorKeychainPath,omitempty"`
        ValidatorRestartCommand string  `yaml:"validatorRestartCommand,omitempty"`
        GasPrice string                 `yaml:"gasPrice,omitempty"`
        GasLimit string                 `yaml:"gasLimit,omitempty"`
    }                                   `yaml:"smartnode,omitempty"`
    Chains struct {
        Eth1 Chain                      `yaml:"eth1,omitempty"`
        Eth2 Chain                      `yaml:"eth2,omitempty"`
    }                                   `yaml:"chains,omitempty"`
}
type Chain struct {
    Provider string                     `yaml:"provider,omitempty"`
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
    Params []ClientParam                `yaml:"params,omitempty"`
}
type ClientParam struct {
    Name string                         `yaml:"name,omitempty"`
    Desc string                         `yaml:"desc,omitempty"`
    Env string                          `yaml:"env,omitempty"`
    Required bool                       `yaml:"required,omitempty"`
    Regex string                        `yaml:"regex,omitempty"`
}
type UserParam struct {
    Env string                          `yaml:"env,omitempty"`
    Value string                        `yaml:"value"`
}


// Get the selected clients from a config
func (config *RocketPoolConfig) GetSelectedEth1Client() *ClientOption {
    return config.Chains.Eth1.GetSelectedClient()
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
    return config, nil
}


// Merge configs
func Merge(configs ...*RocketPoolConfig) RocketPoolConfig {
    var merged RocketPoolConfig
    for i := len(configs) - 1; i >= 0; i-- {
        mergo.Merge(&merged, configs[i])
    }
    return merged
}


// Load merged config from files
func Load(c *cli.Context) (RocketPoolConfig, error) {

    // Load configs
    globalConfig, err := loadFile(os.ExpandEnv(c.GlobalString("config")), true)
    if err != nil {
        return RocketPoolConfig{}, err
    }
    userConfig, err := loadFile(os.ExpandEnv(c.GlobalString("settings")), false)
    if err != nil {
        return RocketPoolConfig{}, err
    }
    cliConfig := getCliConfig(c)

    // Merge and return
    return Merge(&globalConfig, &userConfig, &cliConfig), nil

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
    config.Smartnode.PasswordPath = c.GlobalString("password")
    config.Smartnode.WalletPath = c.GlobalString("wallet")
    config.Smartnode.ValidatorKeychainPath = c.GlobalString("validatorKeychain")
    config.Smartnode.GasPrice = c.GlobalString("gasPrice")
    config.Smartnode.GasLimit = c.GlobalString("gasLimit")
    config.Chains.Eth1.Provider = c.GlobalString("eth1Provider")
    config.Chains.Eth2.Provider = c.GlobalString("eth2Provider")
    return config
}


// Gets gas price from config into big.Int type
func (config *RocketPoolConfig) GetGasPrice() (*big.Int, error) {
    var gasPrice *big.Int
    var err error

    if len(config.Smartnode.GasPrice) > 0 {
        gasPrice, err = cliutils.ValidateWeiAmount("gas price", config.Smartnode.GasPrice)
        if err != nil { 
            return gasPrice, err
        }
        // if 0 detected replace with nil
        if len(gasPrice.Bits()) == 0 { gasPrice = nil } 
    }

    return gasPrice, nil
}


// Gets gas limit from config into uint64 type
func (config *RocketPoolConfig) GetGasLimit() (uint64, error) {
    var gasLimit uint64

    if len(config.Smartnode.GasLimit) > 0 {
        biGasLimit, err := cliutils.ValidateWeiAmount("gas limit", config.Smartnode.GasLimit)
        if err != nil { 
            return gasLimit, err
        }
        gasLimit = biGasLimit.Uint64()
    }

    return gasLimit, nil
}

