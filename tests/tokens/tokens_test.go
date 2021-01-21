package tokens

import (
    "log"
    "math/big"
    "os"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/deposit"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/tests"
    "github.com/rocket-pool/rocketpool-go/tests/utils/accounts"
    "github.com/rocket-pool/rocketpool-go/tests/utils/evm"
    "github.com/rocket-pool/rocketpool-go/tokens"
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


func TestRETHBalances(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Mint rETH
    rethAmount := eth.EthToWei(100)
    if err := mintRETH(userAccount, rethAmount); err != nil { t.Fatal(err) }

    // Get & check rETH total supply
    if rethTotalSupply, err := tokens.GetRETHTotalSupply(rp, nil); err != nil {
        t.Error(err)
    } else if rethTotalSupply.Cmp(rethAmount) != 0 {
        t.Errorf("Incorrect rETH total supply %s", rethTotalSupply.String())
    }

    // Get & check rETH account balance
    if rethBalance, err := tokens.GetRETHBalance(rp, userAccount.Address, nil); err != nil {
        t.Error(err)
    } else if rethBalance.Cmp(rethAmount) != 0 {
        t.Errorf("Incorrect rETH account balance %s", rethBalance.String())
    }

}


// Mint an amount of rETH to an account
func mintRETH(account *accounts.Account, amount *big.Int) error {
    opts := account.GetTransactor()
    opts.Value = amount
    _, err := deposit.Deposit(rp, opts)
    return err
}

