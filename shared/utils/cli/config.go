package cli

import (
    "os"

    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode/shared/utils/config"
)


// Configure the application options
func Configure(app *cli.App) {

    // Register global application options & defaults
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:  "config",
            Usage: "Rocket Pool CLI global config absolute `path`",
            Value: "/.rocketpool/config.yml",
        },
        cli.StringFlag{
            Name:  "settings",
            Usage: "Rocket Pool CLI user config absolute `path`",
            Value: "/.rocketpool/settings.yml",
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
            Name:  "beaconApiMode",
            Usage: "Beacon API mode ('lighthouse' or 'prysm')",
            Value: "",
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

    // Load RP config & set flags
    app.Before = func(c *cli.Context) error {

        // Load config
        _, rpConfig, err := config.Load(c.GlobalString("config"), c.GlobalString("settings"))
        if err != nil { return err }

        // Get selected eth2 client ID
        var eth2ClientId string
        eth2Client := rpConfig.GetSelectedEth2Client()
        if eth2Client != nil { eth2ClientId = eth2Client.ID }

        // Set flags from config
        setFlagFromConfig(c, "database",       rpConfig.Smartnode.DatabasePath)
        setFlagFromConfig(c, "password",       rpConfig.Smartnode.PasswordPath)
        setFlagFromConfig(c, "keychainPow",    rpConfig.Smartnode.NodeKeychainPath)
        setFlagFromConfig(c, "keychainBeacon", rpConfig.Smartnode.ValidatorKeychainPath)
        setFlagFromConfig(c, "providerPow",    rpConfig.Chains.Eth1.Provider)
        setFlagFromConfig(c, "providerBeacon", rpConfig.Chains.Eth2.Provider)
        setFlagFromConfig(c, "storageAddress", rpConfig.Rocketpool.StorageAddress)
        setFlagFromConfig(c, "uniswapAddress", rpConfig.Rocketpool.UniswapAddress)
        setFlagFromConfig(c, "beaconApiMode",  eth2ClientId)

        // Return
        return nil

    }

}


// Set a flag value from a config value
func setFlagFromConfig(c *cli.Context, flagName string, value string) {

    // Cancel if config value is not set
    if value == "" { return }

    // Cancel if flag was set from CLI argument
    for _, arg := range os.Args {
        if arg == "--" + flagName { return }
    }

    // Set flag value
    c.GlobalSet(flagName, value)

}

