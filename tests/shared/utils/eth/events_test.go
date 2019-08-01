package eth

import (
    "bytes"
    "math/big"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// RocketNodeAPI NodeAdd event
type NodeAdd struct {
    ID common.Address
    ContractAddress common.Address
    Created *big.Int
}


// Test getting transaction events
func TestGetTransactionEvents(t *testing.T) {

    // Create account manager & get account
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }
    account, err := am.GetNodeAccount()
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Seed account
    if err := test.SeedAccount(client, account, eth.EthToWei(10)); err != nil { t.Fatal(err) }

    // Initialise contract manager & load test contract
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI"}); err != nil { t.Fatal(err) }

    // Execute test transaction
    txor, err := am.GetNodeAccountTransactor()
    if err != nil { t.Fatal(err) }
    txReceipt, err := eth.ExecuteContractTransaction(client, txor, cm.Addresses["rocketNodeAPI"], cm.Abis["rocketNodeAPI"], "add", "Australia/Brisbane")
    if err != nil { t.Fatal(err) }

    // Get transaction events
    nodeAddEvents, err := eth.GetTransactionEvents(client, txReceipt, cm.Addresses["rocketNodeAPI"], cm.Abis["rocketNodeAPI"], "NodeAdd", NodeAdd{})
    if err != nil {
        t.Fatal(err)
    } else if len(nodeAddEvents) == 0 {
        t.Fatal("Failed to retrieve transaction event")
    }
    nodeAddEvent := (nodeAddEvents[0]).(*NodeAdd)

    // Check node add event values
    if !bytes.Equal(account.Address.Bytes(), nodeAddEvent.ID.Bytes()) { t.Error("Incorrect transaction event values") }

}

