package deposit

import (
    "log"
    "os"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/deposit"
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


func TestDeposit(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Deposit amount
    depositAmount := eth.EthToWei(10)

    // Make deposit
    opts := userAccount.GetTransactor()
    opts.Value = depositAmount
    if _, err := deposit.Deposit(rp, opts); err != nil {
        t.Fatal(err)
    }

    // Get & check deposit pool balance
    if balance, err := deposit.GetBalance(rp, nil); err != nil {
        t.Error(err)
    } else if balance.Cmp(depositAmount) != 0 {
        t.Error("Incorrect deposit pool balance")
    }

    // Get & check deposit pool excess balance
    if excessBalance, err := deposit.GetExcessBalance(rp, nil); err != nil {
        t.Error(err)
    } else if excessBalance.Cmp(depositAmount) != 0 {
        t.Error("Incorrect deposit pool excess balance")
    }

}


func TestAssignDeposits(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Disable deposit assignments
    if _, err := settings.SetAssignDepositsEnabled(rp, false, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Make user deposit
    userDepositOpts := userAccount.GetTransactor()
    userDepositOpts.Value = eth.EthToWei(32)
    if _, err := deposit.Deposit(rp, userDepositOpts); err != nil { t.Fatal(err) }

    // Register node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Make node deposit
    nodeDepositOpts := nodeAccount.GetTransactor()
    nodeDepositOpts.Value = eth.EthToWei(16)
    if _, err := node.Deposit(rp, 0, nodeDepositOpts); err != nil { t.Fatal(err) }

    // Re-enable deposit assignments
    if _, err := settings.SetAssignDepositsEnabled(rp, true, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get initial deposit pool balance
    balance1, err := deposit.GetBalance(rp, nil)
    if err != nil {
        t.Fatal(err)
    }

    // Assign deposits
    if _, err := deposit.AssignDeposits(rp, userAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Get & check updated deposit pool balance
    balance2, err := deposit.GetBalance(rp, nil)
    if err != nil {
        t.Fatal(err)
    } else if balance2.Cmp(balance1) != -1 {
        t.Error("Deposit pool balance did not decrease after assigning deposits")
    }

}

