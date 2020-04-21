package config

import (
    "io/ioutil"

    "github.com/imdario/mergo"
    "gopkg.in/yaml.v2"
)


// Rocket Pool config structure
type RocketPoolConfig struct {
    Rocketpool struct {
        StorageAddress string           `yaml:"storageAddress,omitempty"`
        UniswapAddress string           `yaml:"uniswapAddress,omitempty"`
    }                                   `yaml:"rocketpool,omitempty"`
    Smartnode struct {
        DatabasePath string             `yaml:"databasePath,omitempty"`
        PasswordPath string             `yaml:"passwordPath,omitempty"`
        NodeKeychainPath string         `yaml:"nodeKeychainPath,omitempty"`
        ValidatorKeychainPath string    `yaml:"validatorKeychainPath,omitempty"`
    }                                   `yaml:"smartnode,omitempty"`
    Chains struct {
        Eth1 Chain                      `yaml:"eth1,omitempty"`
        Eth2 Chain                      `yaml:"eth2,omitempty"`
    }                                   `yaml:"chains,omitempty"`
}
type Chain struct {
    Provider string                     `yaml:"provider,omitempty"`
    Client struct {
        Options []ClientOption          `yaml:"options,omitempty"`
        Selected string                 `yaml:"selected,omitempty"`
        Params []UserParam              `yaml:"params,omitempty"`
    }                                   `yaml:"client,omitempty"`
}
type ClientOption struct {
    Name string                         `yaml:"name,omitempty"`
    Image string                        `yaml:"image,omitempty"`
    Params []ClientParam                `yaml:"params,omitempty"`
}
type ClientParam struct {
    Name string                         `yaml:"name,omitempty"`
    Env string                          `yaml:"env,omitempty"`
    Required bool                       `yaml:"required,omitempty"`
    Regex string                        `yaml:"regex,omitempty"`
}
type UserParam struct {
    Env string                          `yaml:"env,omitempty"`
    Value string                        `yaml:"value"`
}


// Config type empty checks
func (c *ClientOption) IsEmpty() bool {
    return c.Name == "" && c.Image == "" && len(c.Params) == 0
}
func (c *ClientParam) IsEmpty() bool {
    return c.Name == "" && c.Env == "" && c.Required == false && c.Regex == ""
}
func (u *UserParam) IsEmpty() bool {
    return u.Env == "" && u.Value == ""
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
        if option.Name == chain.Client.Selected {
            return &option
        }
    }
    return nil
}


// Load Rocket Pool config from config files
// Returns global config and merged config
func Load(globalPath string, userPath string) (*RocketPoolConfig, *RocketPoolConfig, error) {

    // Load config files
    globalConfig, err := loadFile(globalPath)
    if err != nil { return nil, nil, err }
    userConfig, err := loadFile(userPath)
    if err != nil { return nil, nil, err }

    // Merge
    configs := make([]*RocketPoolConfig, 0)
    if globalConfig != nil { configs = append(configs, globalConfig) }
    if userConfig != nil { configs = append(configs, userConfig) }
    mergedConfig := mergeConfigs(configs)

    // Return
    return globalConfig, mergedConfig, nil

}


// Save Rocket Pool config to config files
// Gets diff of merged and global config and saves to user file
func Save(userPath string, globalConfig *RocketPoolConfig, mergedConfig *RocketPoolConfig) error {

    // Diff configs
    if globalConfig != nil && mergedConfig != nil {
        diffConfigs(globalConfig, mergedConfig)
    }

    // Save diff to user config file
    return saveFile(userPath, mergedConfig)

}


// Load Rocket Pool config from a file
func loadFile(path string) (*RocketPoolConfig, error) {

    // Read file
    // Squelch errors due to file not existing; files are optional
    bytes, err := ioutil.ReadFile(path)
    if err != nil { return nil, nil }

    // Parse config
    var config RocketPoolConfig
    if err := yaml.Unmarshal(bytes, &config); err != nil { return nil, err }

    // Return
    return &config, nil

}


// Save Rocket Pool config to a file
func saveFile(path string, config *RocketPoolConfig) error {

    // Encode config
    bytes, err := yaml.Marshal(config)
    if err != nil { return err }

    // Write file
    if err := ioutil.WriteFile(path, bytes, 0644); err != nil { return err }

    // Return
    return nil

}


// Merge configs
func mergeConfigs(configs []*RocketPoolConfig) *RocketPoolConfig {
    var config RocketPoolConfig
    for i := len(configs) - 1; i >= 0; i-- {
        mergo.Merge(&config, configs[i])
    }
    return &config
}


// Diff configs
// Assigns zero value to config B prop if equal to config A prop
func diffConfigs(configA *RocketPoolConfig, configB *RocketPoolConfig) {
    if configA.Rocketpool.StorageAddress       == configB.Rocketpool.StorageAddress       { configB.Rocketpool.StorageAddress = "" }
    if configA.Rocketpool.UniswapAddress       == configB.Rocketpool.UniswapAddress       { configB.Rocketpool.UniswapAddress = "" }
    if configA.Smartnode.DatabasePath          == configB.Smartnode.DatabasePath          { configB.Smartnode.DatabasePath = "" }
    if configA.Smartnode.PasswordPath          == configB.Smartnode.PasswordPath          { configB.Smartnode.PasswordPath = "" }
    if configA.Smartnode.NodeKeychainPath      == configB.Smartnode.NodeKeychainPath      { configB.Smartnode.NodeKeychainPath = "" }
    if configA.Smartnode.ValidatorKeychainPath == configB.Smartnode.ValidatorKeychainPath { configB.Smartnode.ValidatorKeychainPath = "" }
    diffChains(&(configA.Chains.Eth1), &(configB.Chains.Eth1))
    diffChains(&(configA.Chains.Eth2), &(configB.Chains.Eth2))
}
func diffChains(chainA *Chain, chainB *Chain) {
    if chainA.Provider        == chainB.Provider        { chainB.Provider = "" }
    if chainA.Client.Selected == chainB.Client.Selected { chainB.Client.Selected = "" }
    for i := len(chainA.Client.Options) - 1; i >= 0; i-- {
        if i >= len(chainB.Client.Options) { continue }
        diffClientOptions(&(chainA.Client.Options[i]), &(chainB.Client.Options[i]))
        if chainB.Client.Options[i].IsEmpty() { chainB.Client.Options = append(chainB.Client.Options[:i], chainB.Client.Options[i+1:]...) }
    }
    for i := len(chainA.Client.Params) - 1; i >= 0; i-- {
        if i >= len(chainB.Client.Params) { continue }
        diffUserParams(&(chainA.Client.Params[i]), &(chainB.Client.Params[i]))
        if chainB.Client.Params[i].IsEmpty() { chainB.Client.Params = append(chainB.Client.Params[:i], chainB.Client.Params[i+1:]...) }
    }
}
func diffClientOptions(clientA *ClientOption, clientB *ClientOption) {
    if clientA.Name  == clientB.Name  { clientB.Name = "" }
    if clientA.Image == clientB.Image { clientB.Image = "" }
    for i := len(clientA.Params) - 1; i >= 0; i-- {
        if i >= len(clientB.Params) { continue }
        diffClientParams(&(clientA.Params[i]), &(clientB.Params[i]))
        if clientB.Params[i].IsEmpty() { clientB.Params = append(clientB.Params[:i], clientB.Params[i+1:]...) }
    }
}
func diffClientParams(paramA *ClientParam, paramB *ClientParam) {
    if paramA.Name     == paramB.Name     { paramB.Name = "" }
    if paramA.Env      == paramB.Env      { paramB.Env = "" }
    if paramA.Required == paramB.Required { paramB.Required = false }
    if paramA.Regex    == paramB.Regex    { paramB.Regex = "" }
}
func diffUserParams(paramA *UserParam, paramB *UserParam) {
    if paramA.Env   == paramB.Env   { paramB.Env = "" }
    if paramA.Value == paramB.Value { paramB.Value = "" }
}

