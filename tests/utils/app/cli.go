package app

import (
    "bufio"
    "flag"
    "os"
    "regexp"

    "github.com/urfave/cli"

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
    UniswapAddress string
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
        "--uniswapAddress", test.UNISWAP_FACTORY_ADDRESS,
        "--input", inputPath,
        "--output", outputPath,
    }
}


// Get CLI app context
func GetAppContext(dataPath string) *cli.Context {

    // Initialise flag set & define flags
    fs := flag.NewFlagSet("", flag.ContinueOnError)
    fs.String("database", "", "")
    fs.String("password", "", "")
    fs.String("keychainPow", "", "")
    fs.String("keychainBeacon", "", "")
    fs.String("providerPow", "", "")
    fs.String("providerBeacon", "", "")
    fs.String("storageAddress", "", "")
    fs.String("uniswapAddress", "", "")

    // Initialise context
    c := cli.NewContext(nil, fs, nil)

    // Set flags
    c.GlobalSet("database", dataPath + "/rocketpool.db")
    c.GlobalSet("password", dataPath + "/password")
    c.GlobalSet("keychainPow", dataPath + "/accounts")
    c.GlobalSet("keychainBeacon", dataPath + "/validators")
    c.GlobalSet("providerPow", test.POW_PROVIDER_URL)
    c.GlobalSet("providerBeacon", test.BEACON_PROVIDER_URL)
    c.GlobalSet("storageAddress", test.ROCKET_STORAGE_ADDRESS)
    c.GlobalSet("uniswapAddress", test.UNISWAP_FACTORY_ADDRESS)

    // Return context
    return c

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
        UniswapAddress: test.UNISWAP_FACTORY_ADDRESS,
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

