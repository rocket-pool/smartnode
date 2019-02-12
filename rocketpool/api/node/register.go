package node

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/urfave/cli"

    "github.com/rocket-pool/smartnode-cli/rocketpool/utils/eth"
)


// Register the node with Rocket Pool
func registerNode(c *cli.Context) error {

    // Command setup
    am, client, rp, message, err := setup(c, []string{"rocketNodeAPI", "rocketNodeSettings"}, []string{}, true)
    if message != "" {
        fmt.Println(message)
        return nil
    }
    if err != nil {
        return err
    }

    // Check if node is already registered (contract exists)
    nodeContractAddress := new(common.Address)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", am.GetNodeAccount().Address)
    if err != nil {
        return errors.New("Error checking node registration: " + err.Error())
    }
    if !bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
        fmt.Println("Node already registered with contract:", nodeContractAddress.Hex())
        return nil
    }

    // Check node registrations are enabled
    registrationsAllowed := new(bool)
    err = rp.Contracts["rocketNodeSettings"].Call(nil, registrationsAllowed, "getNewAllowed")
    if err != nil {
        return errors.New("Error checking node registrations enabled status: " + err.Error())
    }
    if !*registrationsAllowed {
        fmt.Println("Node registrations are currently disabled in Rocket Pool")
        return nil
    }

    // Get min required node account ether balance
    minNodeAccountEtherBalanceWei := new(*big.Int)
    err = rp.Contracts["rocketNodeSettings"].Call(nil, minNodeAccountEtherBalanceWei, "getEtherMin")
    if err != nil {
        return errors.New("Error retrieving minimum ether requirement: " + err.Error())
    }

    // Check node account ether balance
    nodeAccountEtherBalanceWei, err := client.BalanceAt(context.Background(), am.GetNodeAccount().Address, nil)
    if err != nil {
        return errors.New("Error retrieving node account balance: " + err.Error())
    }
    if nodeAccountEtherBalanceWei.Cmp(*minNodeAccountEtherBalanceWei) < 0 {
        fmt.Println(fmt.Sprintf("Node account requires a minimum balance of %.2f ETH to register", eth.WeiToEth(*minNodeAccountEtherBalanceWei)))
        return nil
    }

    // Prompt user for timezone
    timezone := promptTimezone()

    // Get node account transactor
    nodeAccountTransactor, err := am.GetNodeAccountTransactor()
    if err != nil {
        return err
    }

    // Register node
    _, err = rp.Contracts["rocketNodeAPI"].Transact(nodeAccountTransactor, "add", timezone)
    if err != nil {
        return errors.New("Error registering node: " + err.Error())
    }

    // Get node contract address
    nodeContractAddress = new(common.Address)
    err = rp.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", am.GetNodeAccount().Address)
    if err != nil {
        return errors.New("Error retrieving node contract address: " + err.Error())
    }

    // Log & return
    fmt.Println("Node registered successfully with contract:", nodeContractAddress.Hex())
    return nil

}

