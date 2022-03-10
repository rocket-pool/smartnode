package tokens

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	rplutils "github.com/rocket-pool/rocketpool-go/tests/testutils/tokens/rpl"
)

func TestRPLBalances(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Mint RPL
	rplAmount := eth.EthToWei(100)
	if err := rplutils.MintRPL(rp, ownerAccount, userAccount1, rplAmount); err != nil {
		t.Fatal(err)
	}

	// Get & check RPL account balance
	if rplBalance, err := tokens.GetRPLBalance(rp, userAccount1.Address, nil); err != nil {
		t.Error(err)
	} else if rplBalance.Cmp(rplAmount) != 0 {
		t.Errorf("Incorrect RPL account balance %s", rplBalance.String())
	}

	// Get & check RPL total supply
	initialTotalSupply := eth.EthToWei(18000000)
	if rplTotalSupply, err := tokens.GetRPLTotalSupply(rp, nil); err != nil {
		t.Error(err)
	} else if rplTotalSupply.Cmp(initialTotalSupply) != 0 {
		t.Errorf("Incorrect RPL total supply %s", rplTotalSupply.String())
	}

}

func TestTransferRPL(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Mint RPL
	rplAmount := eth.EthToWei(100)
	if err := rplutils.MintRPL(rp, ownerAccount, userAccount1, rplAmount); err != nil {
		t.Fatal(err)
	}

	// Transfer RPL
	toAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	sendAmount := eth.EthToWei(50)
	if _, err := tokens.TransferRPL(rp, toAddress, sendAmount, userAccount1.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check RPL account balance
	if rplBalance, err := tokens.GetRPLBalance(rp, toAddress, nil); err != nil {
		t.Error(err)
	} else if rplBalance.Cmp(sendAmount) != 0 {
		t.Errorf("Incorrect RPL account balance %s", rplBalance.String())
	}

}

func TestTransferFromRPL(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Mint RPL
	rplAmount := eth.EthToWei(100)
	if err := rplutils.MintRPL(rp, ownerAccount, userAccount1, rplAmount); err != nil {
		t.Fatal(err)
	}

	// Approve RPL spender
	sendAmount := eth.EthToWei(50)
	if _, err := tokens.ApproveRPL(rp, userAccount2.Address, sendAmount, userAccount1.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check spender allowance
	if allowance, err := tokens.GetRPLAllowance(rp, userAccount1.Address, userAccount2.Address, nil); err != nil {
		t.Error(err)
	} else if allowance.Cmp(sendAmount) != 0 {
		t.Errorf("Incorrect RPL spender allowance %s", allowance.String())
	}

	// Transfer RPL from account
	toAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	if _, err := tokens.TransferFromRPL(rp, userAccount1.Address, toAddress, sendAmount, userAccount2.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check RPL account balance
	if rplBalance, err := tokens.GetRPLBalance(rp, toAddress, nil); err != nil {
		t.Error(err)
	} else if rplBalance.Cmp(sendAmount) != 0 {
		t.Errorf("Incorrect RPL account balance %s", rplBalance.String())
	}

}

func TestMintInflationRPL(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Constants
	oneDay := 24 * 60 * 60

	// Start RPL inflation
	if _, err := protocol.BootstrapInflationStartTime(rp, uint64(time.Now().Unix()+3600), ownerAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Increase time until rewards are available
	if err := evm.IncreaseTime(3600 + oneDay); err != nil {
		t.Fatal(err)
	}

	// Get initial total supply
	rplTotalSupply1, err := tokens.GetRPLTotalSupply(rp, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Mint RPL from inflation
	if _, err := tokens.MintInflationRPL(rp, userAccount1.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check updated total supply
	rplTotalSupply2, err := tokens.GetRPLTotalSupply(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rplTotalSupply2.Cmp(rplTotalSupply1) != 1 {
		t.Errorf("Incorrect updated RPL total supply %s", rplTotalSupply2.String())
	}

}

func TestSwapFixedSupplyRPLForRPL(t *testing.T) {

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
	rplAmount := eth.EthToWei(100)
	if err := rplutils.MintFixedSupplyRPL(rp, ownerAccount, userAccount1, rplAmount); err != nil {
		t.Fatal(err)
	}

	// Approve fixed-supply RPL spend
	rocketTokenRPLAddress, err := rp.GetAddress("rocketTokenRPL")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tokens.ApproveFixedSupplyRPL(rp, *rocketTokenRPLAddress, rplAmount, userAccount1.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Swap fixed-supply RP for RPL
	if _, err := tokens.SwapFixedSupplyRPLForRPL(rp, rplAmount, userAccount1.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get & check RPL account balance
	if rplBalance, err := tokens.GetRPLBalance(rp, userAccount1.Address, nil); err != nil {
		t.Error(err)
	} else if rplBalance.Cmp(rplAmount) != 0 {
		t.Errorf("Incorrect RPL account balance %s", rplBalance.String())
	}

}
