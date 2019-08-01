package node

import (
    "testing"

    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/smartnode/shared/services/rocketpool"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/node"
    "github.com/rocket-pool/smartnode/shared/utils/eth"

    test "github.com/rocket-pool/smartnode/tests/utils"
    rp "github.com/rocket-pool/smartnode/tests/utils/rocketpool"
)


// Test node account balances getter
func TestGetAccountBalances(t *testing.T) {

    // Create account manager & get account
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }
    account, err := am.GetNodeAccount()
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketETHToken", "rocketPoolToken"}); err != nil { t.Fatal(err) }

    // Amounts to seed to account
    ethAmount := eth.EthToWei(3)
    rethAmount := eth.EthToWei(0)
    rplAmount := eth.EthToWei(2)

    // Seed account
    if err := test.SeedAccount(client, account.Address, ethAmount); err != nil { t.Fatal(err) }
    if err := rp.MintRPL(client, cm, account.Address, rplAmount); err != nil { t.Fatal(err) }

    // Get account balances
    balances, err := node.GetAccountBalances(account.Address, client, cm)
    if err != nil { t.Fatal(err) }

    // Check account balances
    if balances.EtherWei.String() != ethAmount.String() { t.Errorf("Incorrect balance ETH value: expected %s, got %s", ethAmount.String(), balances.EtherWei.String()) }
    if balances.RethWei.String() != rethAmount.String() { t.Errorf("Incorrect balance rETH value: expected %s, got %s", rethAmount.String(), balances.RethWei.String()) }
    if balances.RplWei.String() != rplAmount.String() { t.Errorf("Incorrect balance RPL value: expected %s, got %s", rplAmount.String(), balances.RplWei.String()) }

}


// Test node contract balances getter
func TestGetBalances(t *testing.T) {

    // Create account manager
    am, err := test.NewInitAccountManager("foobarbaz")
    if err != nil { t.Fatal(err) }

    // Initialise ethereum client
    client, err := ethclient.Dial(test.POW_PROVIDER_URL)
    if err != nil { t.Fatal(err) }

    // Initialise contract manager & load contracts / ABIs
    cm, err := rocketpool.NewContractManager(client, test.ROCKET_STORAGE_ADDRESS)
    if err != nil { t.Fatal(err) }
    if err := cm.LoadContracts([]string{"rocketNodeAPI", "rocketPoolToken"}); err != nil { t.Fatal(err) }
    if err := cm.LoadABIs([]string{"rocketNodeContract"}); err != nil { t.Fatal(err) }

    // Register node
    nodeContract, nodeContractAddress, err := rp.RegisterNode(client, cm, am)
    if err != nil { t.Fatal(err) }

    // Amounts to seed to contract
    ethAmount := eth.EthToWei(5)
    rplAmount := eth.EthToWei(4)

    // Seed contract
    if err := test.SeedAccount(client, nodeContractAddress, ethAmount); err != nil { t.Fatal(err) }
    if err := rp.MintRPL(client, cm, nodeContractAddress, rplAmount); err != nil { t.Fatal(err) }

    // Get contract balances
    balances, err := node.GetBalances(nodeContract)
    if err != nil { t.Fatal(err) }

    // Check contract balances
    if balances.EtherWei.String() != ethAmount.String() { t.Errorf("Incorrect balance ETH value: expected %s, got %s", ethAmount.String(), balances.EtherWei.String()) }
    if balances.RplWei.String() != rplAmount.String() { t.Errorf("Incorrect balance RPL value: expected %s, got %s", rplAmount.String(), balances.RplWei.String()) }

}

