package node

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "math/big"
    "os"
    "os/exec"
    "regexp"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/accounts/keystore"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/services/rocketpool"
    cliutils "github.com/rocket-pool/smartnode-cli/rocketpool/utils/cli"
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
    err = rp.LoadContracts([]string{"rocketNodeAPI", "rocketNodeSettings"})
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

    // Get min required ether balance
    minWeiBalance := new(*big.Int)
    err = rp.Contracts["rocketNodeSettings"].Call(nil, minWeiBalance, "getEtherMin")
    if err != nil {
        return errors.New("Error retrieving minimum ether requirement: " + err.Error())
    }
    var minEtherBalance big.Int
    minEtherBalance.Quo(*minWeiBalance, big.NewInt(1000000000000000000))

    // Check node account balance
    nodeAccountBalance, err := client.BalanceAt(context.Background(), nodeAccount.Address, nil)
    if err != nil {
        return errors.New("Error retrieving node account balance: " + err.Error())
    }
    if nodeAccountBalance.Cmp(*minWeiBalance) < 0 {
        return errors.New(fmt.Sprintf("Node account requires a minimum balance of %s ETH to register", minEtherBalance.String()))
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
        response := cliutils.Prompt(fmt.Sprintf("Your system timezone is '%s', would you like to register using this timezone? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Prompt for time zone
    for timezone == "" {
        timezone = cliutils.Prompt("Please enter a timezone to register with in the format 'Region/City':", "^\\w*\\/\\w*$", "Please enter a timezone in the format 'Region/City'")
        response := cliutils.Prompt(fmt.Sprintf("You have chosen to register with the timezone '%s', is this correct? [y/n]", timezone), "(?i)^(y|yes|n|no)$", "Please answer 'y' or 'n'")
        if strings.ToLower(response[:1]) == "n" {
            timezone = ""
        }
    }

    // Open node account file
    nodeAccountFile, err := os.Open(nodeAccount.URL.Path)
    if err != nil {
        return errors.New("Error opening node account file: " + err.Error())
    }

    // Create node account transactor
    nodeAccountTransactor, err := bind.NewTransactor(nodeAccountFile, "")
    if err != nil {
        return errors.New("Error creating node account transactor: " + err.Error())
    }

    // Register node
    _, err = rp.Contracts["rocketNodeAPI"].Transact(nodeAccountTransactor, "add", timezone)
    if err != nil {
        return errors.New("Error registering node: " + err.Error())
    }

    // Get node contract address
    nodeContractAddress = new(common.Address)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address)
    if err != nil {
        return errors.New("Error retrieving node contract address: " + err.Error())
    }

    // Log & return
    fmt.Println("Node registered successfully with contract:", nodeContractAddress.Hex())
    return nil

}

