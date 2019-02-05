package node

import (
    "bytes"
    "errors"
    "fmt"
    "os/exec"
    "regexp"
    "strings"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/accounts"
    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    cliutils "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
)


// Set the node's timezone
func setNodeTimezone(c *cli.Context) error {

    // Initialise account manager
    am := accounts.NewAccountManager(c.GlobalString("keychain"))

    // Get node account
    if !am.NodeAccountExists() {
        fmt.Println("Node account does not exist, please initialize with `rocketpool node init`")
        return nil
    }
    nodeAccount := am.GetNodeAccount()

    // Connect to ethereum node
    client, err := ethclient.Dial(c.GlobalString("provider"))
    if err != nil {
        return errors.New("Error connecting to ethereum node: " + err.Error())
    }

    // Initialise Rocket Pool contract manager
    rp, err := rocketpool.NewContractManager(client, c.GlobalString("storageAddress"))
    if err != nil {
        return err
    }

    // Load Rocket Pool node contracts
    err = rp.LoadContracts([]string{"rocketNodeAPI"})
    if err != nil {
        return err
    }

    // Check node is registered (contract exists)
    nodeContractAddress := new(common.Address)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address)
    if err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    }
    if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        fmt.Println("Node is not registered with Rocket Pool, please register with `rocketpool node register`")
        return nil
    }

    // Prompt user for timezone
    timezone := promptTimezone()

    // Get node account transactor
    nodeAccountTransactor, err := am.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Set node timezone
    _, err = rp.Contracts["rocketNodeAPI"].Transact(nodeAccountTransactor, "setTimezoneLocation", timezone)
    if err != nil {
        return errors.New("Error setting node timezone: " + err.Error())
    }

    // Get node timezone
    nodeTimezone := new(string)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeTimezone, "getTimezoneLocation", nodeAccount.Address)
    if err != nil {
        return errors.New("Error retrieving node timezone: " + err.Error())
    }

    // Log & return
    fmt.Println("Node timezone successfully updated to:", *nodeTimezone)
    return nil

}


// Prompt user for a time zone string
func promptTimezone() string {

    // Time zone value
    var timezone string

    // Get system time zone
    timeOutput, _ := exec.Command("timedatectl").Output()
    if len(timeOutput) > 0 {
        tzMatches := regexp.MustCompile("(?i)zone:\\s*(\\w{2,}\\/\\w{2,})").FindStringSubmatch(string(timeOutput[:]))
        if len(tzMatches) > 1 {
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

