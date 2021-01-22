package network

import (
    "bytes"
    "log"
    "math/big"
    "os"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/deposit"
    "github.com/rocket-pool/rocketpool-go/network"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/settings"
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
    userAccount *accounts.Account
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
    userAccount, err = accounts.GetAccount(9)
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

    // Get & check network rETH supply
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


func TestNodeFee(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Get settings
    targetNodeFee, err := settings.GetTargetNodeFee(rp, nil)
    if err != nil { t.Fatal(err) }
    minNodeFee, err := settings.GetMinimumNodeFee(rp, nil)
    if err != nil { t.Fatal(err) }
    maxNodeFee, err := settings.GetMaximumNodeFee(rp, nil)
    if err != nil { t.Fatal(err) }
    demandRange, err := settings.GetNodeFeeDemandRange(rp, nil)
    if err != nil { t.Fatal(err) }

    // Get & check initial node demand
    if nodeDemand, err := network.GetNodeDemand(rp, nil); err != nil {
        t.Error(err)
    } else if nodeDemand.Cmp(big.NewInt(0)) != 0 {
        t.Errorf("Incorrect initial node demand value %s", nodeDemand.String())
    }

    // Get & check initial node fee
    if nodeFee, err := network.GetNodeFee(rp, nil); err != nil {
        t.Error(err)
    } else if nodeFee != targetNodeFee {
        t.Errorf("Incorrect initial node fee %f", nodeFee)
    }

    // Make user deposit
    opts := userAccount.GetTransactor()
    opts.Value = demandRange
    if _, err := deposit.Deposit(rp, opts); err != nil { t.Fatal(err) }

    // Get & check updated node demand
    if nodeDemand, err := network.GetNodeDemand(rp, nil); err != nil {
        t.Error(err)
    } else if nodeDemand.Cmp(opts.Value) != 0 {
        t.Errorf("Incorrect updated node demand value %s", nodeDemand.String())
    }

    // Get & check updated node fee
    if nodeFee, err := network.GetNodeFee(rp, nil); err != nil {
        t.Error(err)
    } else if nodeFee != maxNodeFee {
        t.Errorf("Incorrect updated node fee %f", nodeFee)
    }

    // Get & check node fees by demand values
    negDemandRange := new(big.Int)
    negDemandRange.Neg(demandRange)
    if nodeFee, err := network.GetNodeFeeByDemand(rp, big.NewInt(0), nil); err != nil {
        t.Error(err)
    } else if nodeFee != targetNodeFee {
        t.Errorf("Incorrect node fee for zero demand %f", nodeFee)
    }
    if nodeFee, err := network.GetNodeFeeByDemand(rp, negDemandRange, nil); err != nil {
        t.Error(err)
    } else if nodeFee != minNodeFee {
        t.Errorf("Incorrect node fee for negative demand %f", nodeFee)
    }
    if nodeFee, err := network.GetNodeFeeByDemand(rp, demandRange, nil); err != nil {
        t.Error(err)
    } else if nodeFee != maxNodeFee {
        t.Errorf("Incorrect node fee for positive demand %f", nodeFee)
    }

}


func TestSetWithdrawalCredentials(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set withdrawal credentials
    withdrawalCredentials := common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")
    if _, err := network.SetWithdrawalCredentials(rp, withdrawalCredentials, ownerAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check withdrawal credentials
    if networkWithdrawalCredentials, err := network.GetWithdrawalCredentials(rp, nil); err != nil {
        t.Error(err)
    } else if !bytes.Equal(networkWithdrawalCredentials.Bytes(), withdrawalCredentials.Bytes()) {
        t.Errorf("Incorrect network withdrawal credentials %s", networkWithdrawalCredentials.Hex())
    }

}


func TestTransferWithdrawal(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Transfer validator balance
    opts := userAccount.GetTransactor()
    opts.Value = eth.EthToWei(50)
    opts.GasLimit = 100000
    if _, err := network.TransferWithdrawal(rp, opts); err != nil {
        t.Fatal(err)
    }

    // Get & check withdrawal contract balance
    if withdrawalBalance, err := network.GetWithdrawalBalance(rp, nil); err != nil {
        t.Error(err)
    } else if withdrawalBalance.Cmp(opts.Value) != 0 {
        t.Errorf("Incorrect withdrawal contract balance %s", withdrawalBalance.String())
    }

}


func TestProcessWithdrawal(t *testing.T) {
    // TODO: implement
}

