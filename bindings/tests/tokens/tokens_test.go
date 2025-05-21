package tokens

import (
	"math/big"
	"testing"

	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	rethutils "github.com/rocket-pool/rocketpool-go/tests/testutils/tokens/reth"
	rplutils "github.com/rocket-pool/rocketpool-go/tests/testutils/tokens/rpl"
)

func TestTokenBalances(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Mint rETH
	rethAmount := eth.EthToWei(102)
	if err := rethutils.MintRETH(rp, userAccount1, rethAmount); err != nil {
		t.Fatal(err)
	}

	// Mint RPL
	rplAmount := eth.EthToWei(103)
	if err := rplutils.MintRPL(rp, ownerAccount, userAccount1, rplAmount); err != nil {
		t.Fatal(err)
	}

	// Mint fixed-supply RPL
	fixedRplAmount := eth.EthToWei(104)
	if err := rplutils.MintFixedSupplyRPL(rp, ownerAccount, userAccount1, fixedRplAmount); err != nil {
		t.Fatal(err)
	}

	// Get & check token balances
	if balances, err := tokens.GetBalances(rp, userAccount1.Address, nil); err != nil {
		t.Error(err)
	} else {
		if balances.ETH.Cmp(big.NewInt(0)) != 1 {
			t.Errorf("Incorrect ETH balance %s", balances.ETH.String())
		}
		if balances.RETH.Cmp(rethAmount) != 0 {
			t.Errorf("Incorrect rETH balance %s", balances.RETH.String())
		}
		if balances.RPL.Cmp(rplAmount) != 0 {
			t.Errorf("Incorrect RPL balance %s", balances.RPL.String())
		}
		if balances.FixedSupplyRPL.Cmp(fixedRplAmount) != 0 {
			t.Errorf("Incorrect fixed-supply RPL balance %s", balances.FixedSupplyRPL.String())
		}
	}

}
