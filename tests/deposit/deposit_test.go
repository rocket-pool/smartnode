package deposit

import (
    "log"
    "os"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/deposit"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/tests"
    "github.com/rocket-pool/rocketpool-go/tests/utils/accounts"
    "github.com/rocket-pool/rocketpool-go/tests/utils/evm"
)


var (
    client *ethclient.Client
    rp *rocketpool.RocketPool

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
    userAccount, err = accounts.GetAccount(9)
    if err != nil { log.Fatal(err) }

    // Run tests
    os.Exit(m.Run())

}


func TestGetBalance(t *testing.T) {

    // Make staker deposit
    // TODO: implement

    // Get deposit pool balance
    balance, err := deposit.GetBalance(rp, nil)
    if err != nil {
        t.Fatal(err)
    }

    // Check deposit pool balance
    // TODO: implement
    _ = balance

}


func TestAssignDeposits(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Make staker & node deposits
    // TODO: implement

    // Get initial deposit pool balance
    balance1, err := deposit.GetBalance(rp, nil)
    if err != nil {
        t.Fatal(err)
    }

    // Assign deposits
    if _, err := deposit.AssignDeposits(rp, userAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Check updated deposit pool balance
    balance2, err := deposit.GetBalance(rp, nil)
    if err != nil {
        t.Fatal(err)
    } else if balance2.Cmp(balance1) != -1 {
        t.Error("Deposit pool balance did not decrease after assigning deposits")
    }

}

