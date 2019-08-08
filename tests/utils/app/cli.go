package app

import (
    "bufio"
    "os"
    "regexp"

    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode/rocketpool-cli/deposit"
    "github.com/rocket-pool/smartnode/rocketpool-cli/fee"
    "github.com/rocket-pool/smartnode/rocketpool-cli/minipool"
    "github.com/rocket-pool/smartnode/rocketpool-cli/node"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// Application options
type AppOptions struct {
    Database string
    Password string
    KeychainPow string
    KeychainBeacon string
    ProviderPow string
    ProviderBeacon string
    StorageAddress string
}


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
func GetAppArgs(dataPath string, inputPath string, outputPath string) []string {
    return []string{
        "rocketpool-cli",
        "--database", dataPath + "/rocketpool.db",
        "--password", dataPath + "/password",
        "--keychainPow", dataPath + "/accounts",
        "--keychainBeacon", dataPath + "/validators",
        "--providerPow", test.POW_PROVIDER_URL,
        "--providerBeacon", test.BEACON_PROVIDER_URL,
        "--storageAddress", test.ROCKET_STORAGE_ADDRESS,
        "--input", inputPath,
        "--output", outputPath,
    }
}


// Get CLI app options
func GetAppOptions(dataPath string) AppOptions {
    return AppOptions{
        Database: dataPath + "/rocketpool.db",
        Password: dataPath + "/password",
        KeychainPow: dataPath + "/accounts",
        KeychainBeacon: dataPath + "/validators",
        ProviderPow: test.POW_PROVIDER_URL,
        ProviderBeacon: test.BEACON_PROVIDER_URL,
        StorageAddress: test.ROCKET_STORAGE_ADDRESS,
    }
}


// Check output from file against validation rules and return error messages
func CheckOutput(outputPath string, skipLines []string, rules map[int][]string) ([]string, error) {

    // Error messages
    messages := []string{}

    // Open output file
    output, err := os.Open(outputPath)
    if err != nil { return []string{}, err }

    // Scan output
    line := 0
    for scanner := bufio.NewScanner(output); scanner.Scan(); {

        // Check if line should be skipped
        skip := false
        for _, skipLine := range skipLines {
            if regexp.MustCompile(skipLine).MatchString(scanner.Text()) {
                skip = true
                break
            }
        }
        if skip { continue }

        // Increment line number and check line if rules exist
        line++
        if rule, ok := rules[line]; ok {
            if !regexp.MustCompile(rule[0]).MatchString(scanner.Text()) { messages = append(messages, rule[1]) }
        }

    }

    // Check final line count
    if line != len(rules) { messages = append(messages, "Incorrect output line count") }

    // Return error messages
    return messages, nil

}

