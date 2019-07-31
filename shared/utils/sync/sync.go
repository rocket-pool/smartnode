package sync

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/accounts"
    "github.com/rocket-pool/smartnode/shared/services/passwords"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
)


// Config
const CHECK_PASSWORD_INTERVAL string = "10s"
const CHECK_NODE_ACCOUNT_INTERVAL string = "10s"
const CHECK_CONNECTION_INTERVAL string = "10s"
const CHECK_CONTRACT_INTERVAL string = "10s"
const CHECK_NODE_REGISTERED_INTERVAL string = "10s"
var checkPasswordInterval, _ = time.ParseDuration(CHECK_PASSWORD_INTERVAL)
var checkNodeAccountInterval, _ = time.ParseDuration(CHECK_NODE_ACCOUNT_INTERVAL)
var checkConnectionInterval, _ = time.ParseDuration(CHECK_CONNECTION_INTERVAL)
var checkContractInterval, _ = time.ParseDuration(CHECK_CONTRACT_INTERVAL)
var checkNodeRegisteredInterval, _ = time.ParseDuration(CHECK_NODE_REGISTERED_INTERVAL)


// Wait for a password to be set
func WaitPasswordSet(pm *passwords.PasswordManager) {

    // Block until password is set
    for !pm.PasswordExists() {
        fmt.Println(fmt.Sprintf("Node password is not set, retrying in %s...", checkPasswordInterval.String()))
        time.Sleep(checkPasswordInterval)
    }

}


// Wait for a node account to be set
func WaitNodeAccountSet(am *accounts.AccountManager) {

    // Block until node account is set
    for !am.NodeAccountExists() {
        fmt.Println(fmt.Sprintf("Node account does not exist, retrying in %s...", checkNodeAccountInterval.String()))
        time.Sleep(checkNodeAccountInterval)
    }

}


// Wait for ethereum client connection
func WaitClientConnection(client *ethclient.Client) {

    // Block until connected
    var connected bool = false
    for !connected {

        // Get network ID
        if _, err := client.NetworkID(context.Background()); err != nil {
            fmt.Println(fmt.Sprintf("Not connected to ethereum client, retrying in %s...", checkConnectionInterval.String()))
            time.Sleep(checkConnectionInterval)
        } else {
            connected = true
        }

    }

}


// Wait for contract to become available on ethereum client
func WaitContractLoaded(client *ethclient.Client, contractName string, contractAddress common.Address) {

    // Block until contract exists
    var exists bool = false
    for !exists {

        // Get contract code
        if code, err := client.CodeAt(context.Background(), contractAddress, nil); err != nil || len(code) == 0 {
            fmt.Println(fmt.Sprintf("%s contract not loaded, retrying in %s...", contractName, checkContractInterval.String()))
            time.Sleep(checkContractInterval)
        } else {
            exists = true
        }

    }

}


// Wait for node to be registered
func WaitNodeRegistered(am *accounts.AccountManager, cm *rocketpool.ContractManager) error {

    // Block until registered
    var registered bool = false
    for !registered {

        // Get node contract address
        nodeAccount, _ := am.GetNodeAccount()
        nodeContractAddress := new(common.Address)
        if err := cm.Contracts["rocketNodeAPI"].Call(nil, nodeContractAddress, "getContract", nodeAccount.Address); err != nil {
            return errors.New("Error checking node registration: " + err.Error())
        } else if bytes.Equal(nodeContractAddress.Bytes(), make([]byte, common.AddressLength)) {
            fmt.Println(fmt.Sprintf("Node is not registered with Rocket Pool, retrying in %s...", checkNodeRegisteredInterval.String()))
            time.Sleep(checkNodeRegisteredInterval)
        } else {
            registered = true
        }

    }

    // Return
    return nil

}

