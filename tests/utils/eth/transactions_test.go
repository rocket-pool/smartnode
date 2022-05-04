package eth

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/rocketpool-go/tests"
	"github.com/rocket-pool/rocketpool-go/tests/testutils/accounts"
	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	"github.com/rocket-pool/rocketpool-go/utils"
)

func TestSendTransaction(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Initialize eth client
	client, err := ethclient.Dial(tests.Eth1ProviderAddress)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize accounts
	userAccount, err := accounts.GetAccount(9)
	if err != nil {
		t.Fatal(err)
	}

	// Transaction parameters
	toAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	sendAmount := eth.EthToWei(50)

	// Send transaction
	opts := userAccount.GetTransactor()
	opts.Value = sendAmount
	hash, err := eth.SendTransaction(client, toAddress, big.NewInt(1337), opts) // Ganache's default chain ID is 1337
	if err != nil {
		t.Fatal(err)
	}
	if _, err := utils.WaitForTransaction(client, hash); err != nil {
		t.Fatal(err)
	}

	// Get & check to address balance
	if balance, err := client.BalanceAt(context.Background(), toAddress, nil); err != nil {
		t.Error(err)
	} else if balance.Cmp(sendAmount) != 0 {
		t.Errorf("Incorrect to address balance %s", balance.String())
	}

}
