package config

import (
    "io/ioutil"
    "path/filepath"

    "github.com/imdario/mergo"
    "gopkg.in/yaml.v2"
)


// Settings
const GLOBAL_CONFIG_FILENAME = "config.yml"
const USER_CONFIG_FILENAME = "settings.yml"


// Rocket Pool config structure
type RocketPoolConfig struct {
    Rocketpool struct {
        StorageAddress string           `yaml:"storageAddress"`
        UniswapAddress string           `yaml:"uniswapAddress"`
    }                                   `yaml:"rocketpool"`
    Smartnode struct {
        DatabasePath string             `yaml:"databasePath"`
        PasswordPath string             `yaml:"passwordPath"`
        NodeKeychainPath string         `yaml:"nodeKeychainPath"`
        ValidatorKeychainPath string    `yaml:"validatorKeychainPath"`
    }                                   `yaml:"smartnode"`
    Chains struct {
        Eth1 struct {
            Provider string             `yaml:"provider"`
            Client struct {
                Options []ClientOption  `yaml:"options"`
                Selected string         `yaml:"selected"`
            }                           `yaml:"client"`
        }                               `yaml:"eth1"`
        Eth2 struct {
            Provider string             `yaml:"provider"`
            Client struct {
                Options []ClientOption  `yaml:"options"`
                Selected string         `yaml:"selected"`
            }                           `yaml:"client"`
        }                               `yaml:"eth2"`
    }                                   `yaml:"chains"`
}
type ClientOption struct {
    Name string     `yaml:"name"`
    Image string    `yaml:"image"`
}


// Load merged Rocket Pool config from config files
func Load(path string) (*RocketPoolConfig, error) {

    // Config file paths
    filePaths := []string{
        filepath.Join(path, GLOBAL_CONFIG_FILENAME),
        filepath.Join(path, USER_CONFIG_FILENAME),
    }

    // Load configs
    configs := make([]*RocketPoolConfig, 0)
    for _, filePath := range filePaths {
        if config, err := loadFile(filePath); err != nil {
            return nil, err
        } else if config != nil {
            configs = append(configs, config)
        }
    }

    // Merge & return
    config := mergeConfigs(configs)
    return config, nil

}


// Save Rocket Pool config to user config file
func Save(path string, config *RocketPoolConfig) error {

    // Config file path
    filePath := filepath.Join(path, USER_CONFIG_FILENAME)

    // Encode config
    fileContents, err := yaml.Marshal(config)
    if err != nil { return err }

    // Write file
    if err := ioutil.WriteFile(filePath, fileContents, 0644); err != nil { return err }

    // Return
    return nil

}


// Load Rocket Pool config from a file
func loadFile(path string) (*RocketPoolConfig, error) {

    // Read file
    // Squelch errors due to file not existing; config is optional
    bytes, err := ioutil.ReadFile(path)
    if err != nil { return nil, nil }

    // Parse
    var config RocketPoolConfig
    if err := yaml.Unmarshal(bytes, &config); err != nil { return nil, err }

    // Return
    return &config, nil

}


// Merge configs
func mergeConfigs(configs []*RocketPoolConfig) *RocketPoolConfig {
    var config RocketPoolConfig
    for i := len(configs) - 1; i >= 0; i-- {
        mergo.Merge(&config, configs[i])
    }
    return &config
}

