package utils

import (
    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode/rocketpool-cli/deposit"
    "github.com/rocket-pool/smartnode/rocketpool-cli/fee"
    "github.com/rocket-pool/smartnode/rocketpool-cli/minipool"
    "github.com/rocket-pool/smartnode/rocketpool-cli/node"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
)


// Create a test app
func NewApp() *cli.App {

    // Create app
    app := cli.NewApp()

    // Configure
    cliutils.Configure(app)

    // Register commands
    deposit.RegisterCommands(app, "deposit", []string{"d"})
    fee.RegisterCommands(app, "fee", []string{"f"})
    minipool.RegisterCommands(app, "minipool", []string{"m"})
    node.RegisterCommands(app, "node", []string{"n"})

    // Return
    return app

}


// Get CLI app args
func AppArgs(inputPath string, outputPath string, dataPath string) []string {
    return []string{
        "rocketpool-cli",
        "--database", dataPath + "/rocketpool.db",
        "--password", dataPath + "/password",
        "--keychainPow", dataPath + "/accounts",
        "--keychainBeacon", dataPath + "/validators",
        "--providerPow", POW_PROVIDER_URL,
        "--providerBeacon", BEACON_PROVIDER_URL,
        "--storageAddress", ROCKET_STORAGE_ADDRESS,
        "--input", inputPath,
        "--output", outputPath,
    }
}

