package utils

import (
    "bufio"
    "math/big"
    "os"
    "regexp"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/ethclient"
    "gopkg.in/urfave/cli.v1"

    "github.com/rocket-pool/smartnode/rocketpool-cli/deposit"
    "github.com/rocket-pool/smartnode/rocketpool-cli/fee"
    "github.com/rocket-pool/smartnode/rocketpool-cli/minipool"
    "github.com/rocket-pool/smartnode/rocketpool-cli/node"
    "github.com/rocket-pool/smartnode/shared/services/accounts"
    "github.com/rocket-pool/smartnode/shared/services/passwords"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    cliutils "github.com/rocket-pool/smartnode/shared/utils/cli"
    "github.com/rocket-pool/smartnode/shared/utils/eth"
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
        "--providerPow", POW_PROVIDER_URL,
        "--providerBeacon", BEACON_PROVIDER_URL,
        "--storageAddress", ROCKET_STORAGE_ADDRESS,
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
        ProviderPow: POW_PROVIDER_URL,
        ProviderBeacon: BEACON_PROVIDER_URL,
        StorageAddress: ROCKET_STORAGE_ADDRESS,
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


// Seed a node account from app options
func AppSeedNodeAccount(options AppOptions, amount *big.Int) error {

    // Create password manager & account manager
    pm := passwords.NewPasswordManager(nil, nil, options.Password)
    am := accounts.NewAccountManager(options.KeychainPow, pm)

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil { return err }

    // Initialise ethereum client
    client, err := ethclient.Dial(options.ProviderPow)
    if err != nil { return err }

    // Seed account
    return SeedAccount(client, nodeAccount.Address, amount)

}


// Make a node trusted from app options
func AppSetNodeTrusted(options AppOptions) error {

    // Create password manager & account manager
    pm := passwords.NewPasswordManager(nil, nil, options.Password)
    am := accounts.NewAccountManager(options.KeychainPow, pm)

    // Get node account
    nodeAccount, err := am.GetNodeAccount()
    if err != nil { return err }

    // Initialise ethereum client
    client, err := ethclient.Dial(options.ProviderPow)
    if err != nil { return err }

    // Initialise contract manager & load contracts
    cm, err := rocketpool.NewContractManager(client, options.StorageAddress)
    if err != nil { return err }
    if err := cm.LoadContracts([]string{"rocketAdmin"}); err != nil { return err }

    // Get owner account
    ownerPrivateKey, _, err := OwnerAccount()
    if err != nil { return err }

    // Set node trusted status
    txor := bind.NewKeyedTransactor(ownerPrivateKey)
    if _, err := eth.ExecuteContractTransaction(client, txor, cm.Addresses["rocketAdmin"], cm.Abis["rocketAdmin"], "setNodeTrusted", nodeAccount.Address, true); err != nil { return err }

    // Return
    return nil

}

