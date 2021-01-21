package network

import (
    "log"
    "os"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/tests"
    "github.com/rocket-pool/rocketpool-go/tests/utils/accounts"
    "github.com/rocket-pool/rocketpool-go/tests/utils/evm"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
)


var (
    client *ethclient.Client
    rp *rocketpool.RocketPool

    ownerAccount *accounts.Account
    nodeAccount *accounts.Account
)


func TestMain(m *testing.M) {
    var err error

    // Initialize eth client
    client, err = ethclient.Dial(tests.Eth1ProviderAddress)
    if err != nil { log.Fatal(err) }

    // Initialize contract manager
    rp, err = rocketpool.NewRocketPool(client, common.HexToAddress(tests.RocketStorageAddress))
    if err != nil { log.Fatal(err) }

    // Initialize accounts
    ownerAccount, err = accounts.GetAccount(0)
    if err != nil { log.Fatal(err) }
    nodeAccount, err = accounts.GetAccount(1)
    if err != nil { log.Fatal(err) }

    // Run tests
    os.Exit(m.Run())

}


func TestSubmitBalances(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register trusted node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if _, err := node.SetNodeTrusted(rp, nodeAccount.Address, true, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Submit balances
    var balancesBlock uint64 = 100
    totalEth := eth.EthToWei(100)
    stakingEth := eth.EthToWei(80)
    rethSupply := eth.EthToWei(70)
    if _, err := network.SubmitBalances(rp, balancesBlock, totalEth, stakingEth, rethSupply, nodeAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check network balances block
    if networkBalancesBlock, err := network.GetBalancesBlock(rp, nil); err != nil {
        t.Error(err)
    } else if networkBalancesBlock != balancesBlock {
        t.Errorf("Incorrect network balances block %d", networkBalancesBlock)
    }

    // Get & check network total ETH
    if networkTotalEth, err := network.GetTotalETHBalance(rp, nil); err != nil {
        t.Error(err)
    } else if networkTotalEth.Cmp(totalEth) != 0 {
        t.Errorf("Incorrect network total ETH balance %s", networkTotalEth.String())
    }

    // Get & check network staking ETH
    if networkStakingEth, err := network.GetStakingETHBalance(rp, nil); err != nil {
        t.Error(err)
    } else if networkStakingEth.Cmp(stakingEth) != 0 {
        t.Errorf("Incorrect network staking ETH balance %s", networkStakingEth.String())
    }

    // Get & check network
    if networkRethSupply, err := network.GetTotalRETHSupply(rp, nil); err != nil {
        t.Error(err)
    } else if networkRethSupply.Cmp(rethSupply) != 0 {
        t.Errorf("Incorrect network total rETH supply %s", networkRethSupply.String())
    }

    // Get & check ETH utilization rate
    if ethUtilizationRate, err := network.GetETHUtilizationRate(rp, nil); err != nil {
        t.Error(err)
    } else if ethUtilizationRate != eth.WeiToEth(stakingEth) / eth.WeiToEth(totalEth) {
        t.Errorf("Incorrect network ETH utilization rate %f", ethUtilizationRate)
    }

}

