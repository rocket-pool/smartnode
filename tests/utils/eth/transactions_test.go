package eth

import (
    "context"
    "testing"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"

    "github.com/rocket-pool/rocketpool-go/utils/eth"

    "github.com/rocket-pool/rocketpool-go/tests"
    "github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
)


func TestSendTransaction(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Initialize eth client
    client, err := ethclient.Dial(tests.Eth1ProviderAddress)
    if err != nil { t.Fatal(err) }

    // Initialize accounts
    userAccount, err := accounts.GetAccount(9)
    if err != nil { t.Fatal(err) }

    // Transaction parameters
    toAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
    sendAmount := eth.EthToWei(50)

    // Send transaction
    opts := userAccount.GetTransactor()
    opts.Value = sendAmount
    if _, err := eth.SendTransaction(client, toAddress, opts); err != nil {
        t.Fatal(err)
    }

    // Get & check to address balance
    if balance, err := client.BalanceAt(context.Background(), toAddress, nil); err != nil {
        t.Error(err)
    } else if balance.Cmp(sendAmount) != 0 {
        t.Errorf("Incorrect to address balance %s", balance.String())
    }

}

