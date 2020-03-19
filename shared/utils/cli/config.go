package cli

import (
    "github.com/urfave/cli"
    "github.com/urfave/cli/altsrc"
)


// Configure the application options
func Configure(app *cli.App) {

    // Register global application options & defaults
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:  "config",
            Usage: "Rocket Pool CLI config file absolute `path`",
            Value: "/.rocketpool/config.yml",
        },
        altsrc.NewStringFlag(cli.StringFlag{
            Name:  "database",
            Usage: "Rocket Pool CLI database absolute `path`",
            Value: "/.rocketpool/data/rocketpool.db",
        }),
        altsrc.NewStringFlag(cli.StringFlag{
            Name:  "password",
            Usage: "Rocket Pool CLI keystore password `path`",
            Value: "/.rocketpool/data/password",
        }),
        altsrc.NewStringFlag(cli.StringFlag{
            Name:  "keychainPow",
            Usage: "PoW chain account keychain absolute `path`",
            Value: "/.rocketpool/data/accounts",
        }),
        altsrc.NewStringFlag(cli.StringFlag{
            Name:  "keychainBeacon",
            Usage: "Beacon chain account keychain absolute `path`",
            Value: "/.rocketpool/data/validators",
        }),
        altsrc.NewStringFlag(cli.StringFlag{
            Name:  "providerPow",
            Usage: "PoW chain provider `url`",
            Value: "http://127.0.0.1:8545", // Local node
        }),
        altsrc.NewStringFlag(cli.StringFlag{
            Name:  "providerBeacon",
            Usage: "Beacon chain provider `url`",
            Value: "http://127.0.0.1:5052", // Local node
        }),
        altsrc.NewStringFlag(cli.StringFlag{
            Name:  "storageAddress",
            Usage: "PoW chain Rocket Pool storage contract `address`",
            Value: "0x5709b6E58A390534c81dD8EE0E9E1423b843FF5a", // Goerli
        }),
        altsrc.NewStringFlag(cli.StringFlag{
            Name:  "uniswapAddress",
            Usage: "PoW chain Uniswap factory contract `address`",
            Value: "0x6A603658DD351C65379A6fc9f7DD30742ae8bf3c", // Goerli
        }),
        altsrc.NewStringFlag(cli.StringFlag{
            Name:  "input",
            Usage: "Rocket Pool CLI input file `path` (advanced use only)",
        }),
        altsrc.NewStringFlag(cli.StringFlag{
            Name:  "output",
            Usage: "Rocket Pool CLI output file `path` (advanced use only)",
        }),
    }

    // Load external config, squelch load errors
    app.Before = func(c *cli.Context) error {
        altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewYamlSourceFromFlagFunc("config"))(c)
        return nil
    }

}

