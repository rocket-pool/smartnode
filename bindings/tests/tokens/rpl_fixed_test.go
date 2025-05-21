package tokens

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	rplutils "github.com/rocket-pool/rocketpool-go/tests/testutils/tokens/rpl"
)

func TestFixedSupplyRPLBalances(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Mint fixed-supply RPL
	fixedRplAmount := eth.EthToWei(100)
	if err := rplutils.MintFixedSupplyRPL(rp, ownerAccount, userAccount1, fixedRplAmount); err != nil {
		t.Fatal(err)
	}

	// Get & check fixed-supply RPL total supply
	if fixedRplTotalSupply, err := tokens.GetFixedSupplyRPLTotalSupply(rp, nil); err != nil {
		t.Error(err)
	} else if fixedRplTotalSupply.Cmp(fixedRplAmount) != 0 {
		t.Errorf("Incorrect fixed-supply RPL total supply %s", fixedRplTotalSupply.String())
	}

	// Get & check fixed-supply RPL account balance
	if fixedRplBalance, err := tokens.GetFixedSupplyRPLBalance(rp, userAccount1.Address, nil); err != nil {
		t.Error(err)
	} else if fixedRplBalance.Cmp(fixedRplAmount) != 0 {
		t.Errorf("Incorrect fixed-supply RPL account balance %s", fixedRplBalance.String())
	}

}

func TestTransferFixedSupplyRPL(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Mint fixed-supply RPL
	fixedRplAmount := eth.EthToWei(100)
	if err := rplutils.MintFixedSupplyRPL(rp, ownerAccount, userAccount1, fixedRplAmount); err != nil {
		t.Fatal(err)
	}

	// Transfer fixed-supply RPL
	toAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	sendAmount := eth.EthToWei(50)
	if _, err := tokens.TransferFixedSupplyRPL(rp, toAddress, sendAmount, userAccount1.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check fixed-supply RPL account balance
	if fixedRplBalance, err := tokens.GetFixedSupplyRPLBalance(rp, toAddress, nil); err != nil {
		t.Error(err)
	} else if fixedRplBalance.Cmp(sendAmount) != 0 {
		t.Errorf("Incorrect fixed-supply RPL account balance %s", fixedRplBalance.String())
	}

}

func TestTransferFromFixedSupplyRPL(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Mint fixed-supply RPL
	fixedRplAmount := eth.EthToWei(100)
	if err := rplutils.MintFixedSupplyRPL(rp, ownerAccount, userAccount1, fixedRplAmount); err != nil {
		t.Fatal(err)
	}

	// Approve fixed-supply RPL spender
	sendAmount := eth.EthToWei(50)
	if _, err := tokens.ApproveFixedSupplyRPL(rp, userAccount2.Address, sendAmount, userAccount1.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check spender allowance
	if allowance, err := tokens.GetFixedSupplyRPLAllowance(rp, userAccount1.Address, userAccount2.Address, nil); err != nil {
		t.Error(err)
	} else if allowance.Cmp(sendAmount) != 0 {
		t.Errorf("Incorrect fixed-supply RPL spender allowance %s", allowance.String())
	}

	// Transfer fixed-supply RPL from account
	toAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	if _, err := tokens.TransferFromFixedSupplyRPL(rp, userAccount1.Address, toAddress, sendAmount, userAccount2.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check fixed-supply RPL account balance
	if fixedRplBalance, err := tokens.GetFixedSupplyRPLBalance(rp, toAddress, nil); err != nil {
		t.Error(err)
	} else if fixedRplBalance.Cmp(sendAmount) != 0 {
		t.Errorf("Incorrect fixed-supply RPL account balance %s", fixedRplBalance.String())
	}

}
