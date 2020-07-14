package config

import (
    "flag"
    "io/ioutil"

    "github.com/imdario/mergo"
    "gopkg.in/yaml.v2"
)


// Rocket Pool config
type RocketPoolConfig struct {
    Rocketpool struct {
        StorageAddress string           `yaml:"storageAddress,omitempty"`
    }                                   `yaml:"rocketpool,omitempty"`
    Smartnode struct {
        PasswordPath string             `yaml:"passwordPath,omitempty"`
        NodeKeychainPath string         `yaml:"nodeKeychainPath,omitempty"`
        ValidatorKeychainPath string    `yaml:"validatorKeychainPath,omitempty"`
    }                                   `yaml:"smartnode,omitempty"`
    Chains struct {
        Eth1 chain                      `yaml:"eth1,omitempty"`
        Eth2 chain                      `yaml:"eth2,omitempty"`
    }                                   `yaml:"chains,omitempty"`
}
type chain struct {
    Provider string                     `yaml:"provider,omitempty"`
    Client struct {
        Options []clientOption          `yaml:"options,omitempty"`
        Selected string                 `yaml:"selected,omitempty"`
        Params []userParam              `yaml:"params,omitempty"`
    }                                   `yaml:"client,omitempty"`
}
type clientOption struct {
    ID string                           `yaml:"id,omitempty"`
    Name string                         `yaml:"name,omitempty"`
    Image string                        `yaml:"image,omitempty"`
    Params []clientParam                `yaml:"params,omitempty"`
}
type clientParam struct {
    Name string                         `yaml:"name,omitempty"`
    Env string                          `yaml:"env,omitempty"`
    Required bool                       `yaml:"required,omitempty"`
    Regex string                        `yaml:"regex,omitempty"`
}
type userParam struct {
    Env string                          `yaml:"env,omitempty"`
    Value string                        `yaml:"value"`
}


// Get the selected clients from a config
func (config *RocketPoolConfig) GetSelectedEth1Client() *clientOption {
    return config.Chains.Eth1.GetSelectedClient()
}
func (config *RocketPoolConfig) GetSelectedEth2Client() *clientOption {
    return config.Chains.Eth2.GetSelectedClient()
}
func (chain *chain) GetSelectedClient() *clientOption {
    for _, option := range chain.Client.Options {
        if option.ID == chain.Client.Selected {
            return &option
        }
    }
    return nil
}


// Load Rocket Pool config from files
// Returns global config and merged config
func Load(globalPath string, userPath string) (RocketPoolConfig, RocketPoolConfig, error) {

    // Load config files
    globalConfig, err := loadFile(globalPath)
    if err != nil {
        return RocketPoolConfig{}, RocketPoolConfig{}, err
    }
    userConfig, err := loadFile(userPath)
    if err != nil {
        return RocketPoolConfig{}, RocketPoolConfig{}, err
    }

    // Load CLI config
    cliConfig := loadCliConfig()

    // Merge
    mergedConfig := mergeConfigs(&globalConfig, &userConfig, &cliConfig)

    // Return
    return globalConfig, mergedConfig, nil

}


// Load Rocket Pool config from a file
func loadFile(path string) (RocketPoolConfig, error) {

    // Read file
    // Squelch not found errors; config files are optional
    bytes, err := ioutil.ReadFile(path)
    if err != nil {
        return RocketPoolConfig{}, nil
    }

    // Parse config
    var config RocketPoolConfig
    if err := yaml.Unmarshal(bytes, &config); err != nil {
        return RocketPoolConfig{}, err
    }

    // Return
    return config, nil

}


// Create Rocket Pool config from CLI arguments
func loadCliConfig() RocketPoolConfig {

    // Define & parse flags
    storageAddress :=    flag.String("storageAddress",    "", "Rocket Pool storage contract address")
    password :=          flag.String("password",          "", "Keystore password path")
    nodeKeychain :=      flag.String("nodeKeychain",      "", "Eth 1.0 account keychain path")
    validatorKeychain := flag.String("validatorKeychain", "", "Eth 2.0 account keychain path")
    eth1Provider :=      flag.String("eth1Provider",      "", "Eth 1.0 provider address")
    eth2Provider :=      flag.String("eth2Provider",      "", "Eth 2.0 provider address")
    flag.Parse()

    // Return
    return RocketPoolConfig{
        Rocketpool: {
            StorageAddress: *storageAddress,
        },
        Smartnode: {
            PasswordPath: *password,
            NodeKeychainPath: *nodeKeychain,
            ValidatorKeychainPath: *validatorKeychain,
        },
        Chains: {
            Eth1: {
                Provider: *eth1Provider,
            },
            Eth2: {
                Provider: *eth2Provider,
            },
        },
    }

}


// Merge Rocket Pool configs
func mergeConfigs(configs ...*RocketPoolConfig) RocketPoolConfig {
    var merged RocketPoolConfig
    for i := len(configs) - 1; i >= 0; i-- {
        mergo.Merge(&merged, configs[i])
    }
    return merged
}

