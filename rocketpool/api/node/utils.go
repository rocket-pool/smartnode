package node

import (
    "errors"
    "fmt"
    "os/exec"
    "regexp"
    "strings"

    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    cliutils "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
)


// Shared command vars
var am = new(accounts.AccountManager)
var client = new(ethclient.Client)
var cm = new(rocketpool.ContractManager)


// Shared command setup
func setup(c *cli.Context, loadContracts []string, loadAbis []string, accountRequired bool) (string, error) {

    // Initialise account manager
    *am = *accounts.NewAccountManager(c.GlobalString("keychain"))

    // Check node account
    if !am.NodeAccountExists() {
        if accountRequired {
            return "Node account does not exist, please initialize with `rocketpool node init`", nil
        } else {
            return "Node account has not been initialized", nil
        }
    }

    // Connect to ethereum node
    if clientV, err := ethclient.Dial(c.GlobalString("provider")); err != nil {
        return "", errors.New("Error connecting to ethereum node: " + err.Error())
    } else {
        *client = *clientV
    }

    // Initialise Rocket Pool contract manager
    if cmV, err := rocketpool.NewContractManager(client, c.GlobalString("storageAddress")); err != nil {
        return "", err
    } else {
        *cm = *cmV
    }

    // Loading channels
    successChannel := make(chan bool)
    errorChannel := make(chan error)

    // Load Rocket Pool contracts
    go (func() {
        if err := cm.LoadContracts(loadContracts); err != nil {
            errorChannel <- err
        } else {
            successChannel <- true
        }
    })()
    go (func() {
        if err := cm.LoadABIs(loadAbis); err != nil {
            errorChannel <- err
        } else {
            successChannel <- true
        }
    })()

    // Await loading
    for received := 0; received < 2; {
        select {
            case <-successChannel:
                received++
            case err := <-errorChannel:
                return "", err
        }
    }

    // Return
    return "", nil

}


// Prompt user for a time zone string
func promptTimezone() string {

    // Time zone value
    var timezone string

    // Get system time zone
    if timeOutput, _ := exec.Command("timedatectl").Output(); len(timeOutput) > 0 {
        if tzMatches := regexp.MustCompile("(?i)zone:\\s*(\\w{2,}\\/\\w{2,})").FindStringSubmatch(string(timeOutput[:])); len(tzMatches) > 1 {
            timezone = tzMatches[1]
        }
    }

    // Confirm system time zone
    if timezone != "" {
        response := cliutils.Prompt(fmt.Sprintf("Your system timezone is '%s', would you like to register using this timezone? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Prompt for time zone
    for timezone == "" {
        timezone = cliutils.Prompt("Please enter a timezone to register with in the format 'Country/City':", "^\\w{2,}\\/\\w{2,}$", "Please enter a timezone in the format 'Country/City'")
        response := cliutils.Prompt(fmt.Sprintf("You have chosen to register with the timezone '%s', is this correct? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Return
    return timezone

}

