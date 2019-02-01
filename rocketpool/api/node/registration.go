package node

import (
    "bytes"
    "errors"
    "fmt"
    "os/exec"
    "regexp"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/keystore"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/commands"
)


// Register the node with Rocket Pool
func registerNode(c *cli.Context) error {

    // Initialise keystore
    ks := keystore.NewKeyStore(c.GlobalString("keychain"), keystore.StandardScryptN, keystore.StandardScryptP)

    // Get node account
    if len(ks.Accounts()) == 0 {
        return errors.New("Node account does not exist, please initialize with `rocketpool node init`")
    }
    nodeAccount := ks.Accounts()[0]

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

    // Check if node is already registered (contract exists)
    nodeContractAddress := new(common.Address)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address)
    if err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    }
    if !bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        fmt.Println("Node already registered with contract:", nodeContractAddress.Hex())
        return nil
    }

    // Get system time zone
    var timezone string
    timeOutput, _ := exec.Command("timedatectl").Output()
    if len(timeOutput) > 0 {
        tzMatches := regexp.MustCompile("(?i)zone:\\s*(\\w*\\/\\w*)").FindStringSubmatch(string(timeOutput[:]))
        if len(tzMatches) > 1 {
            timezone = tzMatches[1]
        }
    }

    // Confirm system time zone
    if timezone != "" {
        response := commands.Prompt(fmt.Sprintf("Your system timezone is '%s', would you like to register using this timezone? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Prompt for time zone
    for timezone == "" {
        timezone = commands.Prompt("Please enter a timezone to register with in the format 'Region/City':", "^\\w*\\/\\w*$", "Please enter a timezone in the format 'Region/City'")
        response := commands.Prompt(fmt.Sprintf("You have chosen to register with the timezone '%s', is this correct? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Register node
    

    // Return
    return nil

}

