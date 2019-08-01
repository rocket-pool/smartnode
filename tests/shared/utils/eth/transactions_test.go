package eth

import (
    "testing"

    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
)


// Test executing contract transactions
func TestExecuteContractTransaction(t *testing.T) {

    // Create account manager & get account
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }
    account, err := am.GetNodeAccount()
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Seed account
    if err := test.SeedAccount(client, account.Address, eth.EthToWei(10)); err != nil { t.Fatal(err) }

    // Initialise contract manager & load test contract
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI"}); err != nil { t.Fatal(err) }

    // Execute test transaction
    txor, err := am.GetNodeAccountTransactor()
    if err != nil { t.Fatal(err) }
    if _, err := eth.ExecuteContractTransaction(client, txor, cm.Addresses["rocketNodeAPI"], cm.Abis["rocketNodeAPI"], "add", "Australia/Brisbane"); err != nil { t.Fatal(err) }

}

