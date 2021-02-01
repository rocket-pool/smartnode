package rocketpool

import (
    "log"
    "os"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

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

    // Run tests
    os.Exit(m.Run())

}


func TestDepositSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set & get deposits enabled
    depositEnabled := false
    if _, err := settings.SetDepositEnabled(rp, depositEnabled, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetDepositEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if value != depositEnabled {
        t.Error("Incorrect deposit enabled value")
    }

    // Set & get deposit assignments enabled
    assignDepositsEnabled := false
    if _, err := settings.SetAssignDepositsEnabled(rp, assignDepositsEnabled, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetAssignDepositsEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if value != assignDepositsEnabled {
        t.Error("Incorrect assign deposits enabled value")
    }

    // Set & get minimum deposit amount
    minimumDeposit := eth.EthToWei(1000)
    if _, err := settings.SetMinimumDeposit(rp, minimumDeposit, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetMinimumDeposit(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(minimumDeposit) != 0 {
        t.Error("Incorrect minimum deposit value")
    }

    // Set & get maximum deposit pool size
    maximumDepositPoolSize := eth.EthToWei(1)
    if _, err := settings.SetMaximumDepositPoolSize(rp, maximumDepositPoolSize, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetMaximumDepositPoolSize(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(maximumDepositPoolSize) != 0 {
        t.Error("Incorrect maximum deposit pool size value")
    }

    // Set & get maximum deposit assignments
    var maximumDepositAssignments uint64 = 50
    if _, err := settings.SetMaximumDepositAssignments(rp, maximumDepositAssignments, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetMaximumDepositAssignments(rp, nil); err != nil {
        t.Error(err)
    } else if value != maximumDepositAssignments {
        t.Error("Incorrect maximum deposit assignments value")
    }

}


func TestMinipoolSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Get & check launch balance and deposit amounts
    fullMinipoolBalance := eth.EthToWei(32)
    halfMinipoolBalance := eth.EthToWei(16)
    emptyMinipoolBalance := eth.EthToWei(0)
    if value, err := settings.GetMinipoolLaunchBalance(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(fullMinipoolBalance) != 0 {
        t.Error("Incorrect minipool launch balance")
    }
    if value, err := settings.GetMinipoolFullDepositNodeAmount(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(fullMinipoolBalance) != 0 {
        t.Error("Incorrect minipool full deposit node amount")
    }
    if value, err := settings.GetMinipoolHalfDepositNodeAmount(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(halfMinipoolBalance) != 0 {
        t.Error("Incorrect minipool half deposit node amount")
    }
    if value, err := settings.GetMinipoolEmptyDepositNodeAmount(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(emptyMinipoolBalance) != 0 {
        t.Error("Incorrect minipool empty deposit node amount")
    }
    if value, err := settings.GetMinipoolFullDepositUserAmount(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(halfMinipoolBalance) != 0 {
        t.Error("Incorrect minipool full deposit user amount")
    }
    if value, err := settings.GetMinipoolHalfDepositUserAmount(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(halfMinipoolBalance) != 0 {
        t.Error("Incorrect minipool half deposit user amount")
    }
    if value, err := settings.GetMinipoolEmptyDepositUserAmount(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(fullMinipoolBalance) != 0 {
        t.Error("Incorrect minipool empty deposit user amount")
    }

    // Set & get submit withdrawable enabled
    submitWithdrawableEnabled := false
    if _, err := settings.SetMinipoolSubmitWithdrawableEnabled(rp, submitWithdrawableEnabled, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetMinipoolSubmitWithdrawableEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if value != submitWithdrawableEnabled {
        t.Error("Incorrect minipool withdrawable submissions enabled value")
    }

    // Set & get minipool launch timeout
    var minipoolLaunchTimeout uint64 = 5
    if _, err := settings.SetMinipoolLaunchTimeout(rp, minipoolLaunchTimeout, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetMinipoolLaunchTimeout(rp, nil); err != nil {
        t.Error(err)
    } else if value != minipoolLaunchTimeout {
        t.Error("Incorrect minipool launch timeout value")
    }

    // Set & get minipool withdrawal delay
    var minipoolWithdrawalDelay uint64 = 5
    if _, err := settings.SetMinipoolWithdrawalDelay(rp, minipoolWithdrawalDelay, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetMinipoolWithdrawalDelay(rp, nil); err != nil {
        t.Error(err)
    } else if value != minipoolWithdrawalDelay {
        t.Error("Incorrect minipool withdrawal delay value")
    }

}


func TestNetworkSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set & get node consensus threshold
    nodeConsensusThreshold := 0.1
    if _, err := settings.SetNodeConsensusThreshold(rp, nodeConsensusThreshold, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetNodeConsensusThreshold(rp, nil); err != nil {
        t.Error(err)
    } else if value != nodeConsensusThreshold {
        t.Error("Incorrect node consensus threshold value")
    }

    // Set & get network balance submissions enabled
    submitBalancesEnabled := false
    if _, err := settings.SetSubmitBalancesEnabled(rp, submitBalancesEnabled, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetSubmitBalancesEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if value != submitBalancesEnabled {
        t.Error("Incorrect network balance submissions enabled value")
    }

    // Set & get network balance submission frequency
    var submitBalancesFrequency uint64 = 10
    if _, err := settings.SetSubmitBalancesFrequency(rp, submitBalancesFrequency, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetSubmitBalancesFrequency(rp, nil); err != nil {
        t.Error(err)
    } else if value != submitBalancesFrequency {
        t.Error("Incorrect network balance submission frequency value")
    }

    // Set & get process withdrawals enabled
    processWithdrawalsEnabled := false
    if _, err := settings.SetProcessWithdrawalsEnabled(rp, processWithdrawalsEnabled, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetProcessWithdrawalsEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if value != processWithdrawalsEnabled {
        t.Error("Incorrect process withdrawals enabled value")
    }

    // Set & get minimum node fee
    minimumNodeFee := 0.80
    if _, err := settings.SetMinimumNodeFee(rp, minimumNodeFee, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetMinimumNodeFee(rp, nil); err != nil {
        t.Error(err)
    } else if value != minimumNodeFee {
        t.Error("Incorrect minimum node fee value")
    }

    // Set & get target node fee
    targetNodeFee := 0.85
    if _, err := settings.SetTargetNodeFee(rp, targetNodeFee, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetTargetNodeFee(rp, nil); err != nil {
        t.Error(err)
    } else if value != targetNodeFee {
        t.Error("Incorrect target node fee value")
    }

    // Set & get maximum node fee
    maximumNodeFee := 0.90
    if _, err := settings.SetMaximumNodeFee(rp, maximumNodeFee, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetMaximumNodeFee(rp, nil); err != nil {
        t.Error(err)
    } else if value != maximumNodeFee {
        t.Error("Incorrect maximum node fee value")
    }

    // Set & get node fee demand range
    nodeFeeDemandRange := eth.EthToWei(10)
    if _, err := settings.SetNodeFeeDemandRange(rp, nodeFeeDemandRange, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetNodeFeeDemandRange(rp, nil); err != nil {
        t.Error(err)
    } else if value.Cmp(nodeFeeDemandRange) != 0 {
        t.Error("Incorrect node fee demand range value")
    }

    // Set & get target rETH collateral rate
    targetRethCollateralRate := 0.95
    if _, err := settings.SetTargetRethCollateralRate(rp, targetRethCollateralRate, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetTargetRethCollateralRate(rp, nil); err != nil {
        t.Error(err)
    } else if value != targetRethCollateralRate {
        t.Error("Incorrect target rETH collateral rate value")
    }

}


func TestNodeSettings(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Set & get node registrations enabled
    nodeRegistrationsEnabled := false
    if _, err := settings.SetNodeRegistrationEnabled(rp, nodeRegistrationsEnabled, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetNodeRegistrationEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if value != nodeRegistrationsEnabled {
        t.Error("Incorrect node registrations enabled value")
    }

    // Set & get node deposits enabled
    nodeDepositsEnabled := false
    if _, err := settings.SetNodeDepositEnabled(rp, nodeDepositsEnabled, ownerAccount.GetTransactor()); err != nil {
        t.Error(err)
    } else if value, err := settings.GetNodeDepositEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if value != nodeDepositsEnabled {
        t.Error("Incorrect node deposits enabled value")
    }

}

