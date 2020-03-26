package cli

import (
    "io/ioutil"
    "os"

    "github.com/urfave/cli"
    "gopkg.in/yaml.v2"
)


// Rocket Pool config structure
type RocketPoolConfig struct {
    Rocketpool struct {
        StorageAddress string       `yaml:"storageAddress"`
        UniswapAddress string       `yaml:"uniswapAddress"`
    }                           `yaml:"rocketpool"`
    Chains struct {
        Eth1 struct {
            Provider string             `yaml:"provider"`
        }                           `yaml:"eth1"`
        Eth2 struct {
            Provider string             `yaml:"provider"`
        }                           `yaml:"eth2"`
    }                           `yaml:"chains"`
}


// Configure the application options
func Configure(app *cli.App) {

    // Register global application options & defaults
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:  "config",
            Usage: "Rocket Pool CLI config file absolute `path`",
            Value: "/.rocketpool/config.yml",
        },
        cli.StringFlag{
            Name:  "database",
            Usage: "Rocket Pool CLI database absolute `path`",
            Value: "/.rocketpool/data/rocketpool.db",
        },
        cli.StringFlag{
            Name:  "password",
            Usage: "Rocket Pool CLI keystore password `path`",
            Value: "/.rocketpool/data/password",
        },
        cli.StringFlag{
            Name:  "keychainPow",
            Usage: "PoW chain account keychain absolute `path`",
            Value: "/.rocketpool/data/accounts",
        },
        cli.StringFlag{
            Name:  "keychainBeacon",
            Usage: "Beacon chain account keychain absolute `path`",
            Value: "/.rocketpool/data/validators",
        },
        cli.StringFlag{
            Name:  "providerPow",
            Usage: "PoW chain provider `url`",
            Value: "http://127.0.0.1:8545", // Local node
        },
        cli.StringFlag{
            Name:  "providerBeacon",
            Usage: "Beacon chain provider `url`",
            Value: "http://127.0.0.1:5052", // Local node
        },
        cli.StringFlag{
            Name:  "storageAddress",
            Usage: "PoW chain Rocket Pool storage contract `address`",
            Value: "0x5709b6E58A390534c81dD8EE0E9E1423b843FF5a", // Goerli
        },
        cli.StringFlag{
            Name:  "uniswapAddress",
            Usage: "PoW chain Uniswap factory contract `address`",
            Value: "0x6A603658DD351C65379A6fc9f7DD30742ae8bf3c", // Goerli
        },
        cli.StringFlag{
            Name:  "input",
            Usage: "Rocket Pool CLI input file `path` (advanced use only)",
        },
        cli.StringFlag{
            Name:  "output",
            Usage: "Rocket Pool CLI output file `path` (advanced use only)",
        },
    }

    // Load config file & set flags
    app.Before = func(c *cli.Context) error {

        // Load config
        config, err := loadConfigFile(c.GlobalString("config"))
        if err != nil { return err }

        // Set flag values from config
        applyFlagConfig(c, "providerPow",    config.Chains.Eth1.Provider)
        applyFlagConfig(c, "providerBeacon", config.Chains.Eth2.Provider)
        applyFlagConfig(c, "storageAddress", config.Rocketpool.StorageAddress)
        applyFlagConfig(c, "uniswapAddress", config.Rocketpool.UniswapAddress)

        // Return
        return nil

    }

}


// Load and parse a YAML config file
func loadConfigFile(path string) (*RocketPoolConfig, error) {

    // Read file
    bytes, err := ioutil.ReadFile(path)
    if err != nil { return nil, nil } // Squelch errors from config file not existing

    // Parse
    var config RocketPoolConfig
    if err := yaml.Unmarshal(bytes, &config); err != nil { return nil, err }

    // Return
    return &config, nil

}


// Set a flag value from a config value
func applyFlagConfig(c *cli.Context, flagName string, value string) {

    // Cancel if config value is undefined
    if value == "" { return }

    // Cancel if flag was set from CLI argument
    for _, arg := range os.Args {
        if arg == "--" + flagName { return }
    }

    // Set flag value
    c.GlobalSet(flagName, value)

}

