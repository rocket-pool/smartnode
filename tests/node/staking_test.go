package node

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/rocket-pool/rocketpool-go/deposit"
	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/settings/trustednode"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	minipoolutils "github.com/rocket-pool/rocketpool-go/tests/testutils/minipool"
	nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
	rplutils "github.com/rocket-pool/rocketpool-go/tests/testutils/tokens/rpl"
)

func TestStakeRPL(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Register node
	if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Get RPL amount required for 2 minipools
	minipoolRplRequired, err := minipoolutils.GetMinipoolRPLRequired(rp)
	if err != nil {
		t.Fatal(err)
	}
	rplAmount := new(big.Int)
	rplAmount.Mul(minipoolRplRequired, big.NewInt(2))

	// Mint RPL
	if err := rplutils.MintRPL(rp, ownerAccount, nodeAccount, rplAmount); err != nil {
		t.Fatal(err)
	}

	// Approve RPL transfer for staking
	rocketNodeStakingAddress, err := rp.GetAddress("rocketNodeStaking")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tokens.ApproveRPL(rp, *rocketNodeStakingAddress, rplAmount, nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Check initial staking details
	if totalRplStake, err := node.GetTotalRPLStake(rp, nil); err != nil {
		t.Error(err)
	} else if totalRplStake.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect initial total RPL stake %s", totalRplStake.String())
	}
	if totalEffectiveRplStake, err := node.GetTotalEffectiveRPLStake(rp, nil); err != nil {
		t.Error(err)
	} else if totalEffectiveRplStake.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect initial total effective RPL stake %s", totalEffectiveRplStake.String())
	}
	if nodeRplStake, err := node.GetNodeRPLStake(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeRplStake.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect initial node RPL stake %s", nodeRplStake.String())
	}
	if nodeEffectiveRplStake, err := node.GetNodeEffectiveRPLStake(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeEffectiveRplStake.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect initial node effective RPL stake %s", nodeEffectiveRplStake.String())
	}
	if nodeMinimumRplStake, err := node.GetNodeMinimumRPLStake(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeMinimumRplStake.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect initial node minimum RPL stake %s", nodeMinimumRplStake.String())
	}
	if nodeRplStakedTime, err := node.GetNodeRPLStakedTime(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeRplStakedTime != 0 {
		t.Errorf("Incorrect initial node RPL staked time %d", nodeRplStakedTime)
	}
	if nodeMinipoolLimit, err := node.GetNodeMinipoolLimit(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeMinipoolLimit != 0 {
		t.Errorf("Incorrect initial node minipool limit %d", nodeMinipoolLimit)
	}

	// Stake RPL
	if _, err := node.StakeRPL(rp, rplAmount, nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Check updated staking details
	if totalRplStake, err := node.GetTotalRPLStake(rp, nil); err != nil {
		t.Error(err)
	} else if totalRplStake.Cmp(rplAmount) != 0 {
		t.Errorf("Incorrect updated total RPL stake 1 %s", totalRplStake.String())
	}
	if totalEffectiveRplStake, err := node.GetTotalEffectiveRPLStake(rp, nil); err != nil {
		t.Error(err)
	} else if totalEffectiveRplStake.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect updated total effective RPL stake 1 %s", totalEffectiveRplStake.String())
	}
	if nodeRplStake, err := node.GetNodeRPLStake(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeRplStake.Cmp(rplAmount) != 0 {
		t.Errorf("Incorrect updated node RPL stake 1 %s", nodeRplStake.String())
	}
	if nodeEffectiveRplStake, err := node.GetNodeEffectiveRPLStake(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeEffectiveRplStake.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect updated node effective RPL stake 1 %s", nodeEffectiveRplStake.String())
	}
	if nodeMinimumRplStake, err := node.GetNodeMinimumRPLStake(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeMinimumRplStake.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect updated node minimum RPL stake 1 %s", nodeMinimumRplStake.String())
	}
	if nodeRplStakedTime, err := node.GetNodeRPLStakedTime(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeRplStakedTime == 0 {
		t.Errorf("Incorrect updated node RPL staked time 1 %d", nodeRplStakedTime)
	}
	if nodeMinipoolLimit, err := node.GetNodeMinipoolLimit(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeMinipoolLimit != 2 {
		t.Errorf("Incorrect updated node minipool limit 1 %d", nodeMinipoolLimit)
	}

	// Make node deposit to create minipool
	minipoolAddress, _, err := nodeutils.Deposit(t, rp, nodeAccount, eth.EthToWei(16), 1)
	if err != nil {
		t.Fatal(err)
	}
	mp, err := minipool.NewMinipool(rp, minipoolAddress)
	if err != nil {
		t.Fatal(err)
	}

	// Make user deposit
	depositOpts := nodeAccount.GetTransactor()
	depositOpts.Value = eth.EthToWei(16)
	if _, err := deposit.Deposit(rp, depositOpts); err != nil {
		t.Fatal(err)
	}

	// Delay for the time between depositing and staking
	scrubPeriod, err := trustednode.GetScrubPeriod(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = evm.IncreaseTime(int(scrubPeriod + 1))
	if err != nil {
		t.Fatal(fmt.Errorf("Could not increase time: %w", err))
	}

	// Stake minipool
	if err := minipoolutils.StakeMinipool(rp, mp, nodeAccount); err != nil {
		t.Fatal(err)
	}

	// Check updated staking details
	if totalEffectiveRplStake, err := node.GetTotalEffectiveRPLStake(rp, nil); err != nil {
		t.Error(err)
	} else if totalEffectiveRplStake.Cmp(rplAmount) != 0 {
		t.Errorf("Incorrect updated total effective RPL stake 2 %s", totalEffectiveRplStake.String())
	}
	if nodeEffectiveRplStake, err := node.GetNodeEffectiveRPLStake(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeEffectiveRplStake.Cmp(rplAmount) != 0 {
		t.Errorf("Incorrect updated node effective RPL stake 2 %s", nodeEffectiveRplStake.String())
	}
	if nodeMinimumRplStake, err := node.GetNodeMinimumRPLStake(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeMinimumRplStake.Cmp(minipoolRplRequired) != 0 {
		t.Errorf("Incorrect updated node minimum RPL stake 2 %s", nodeMinimumRplStake.String())
	}

}

func TestWithdrawRPL(t *testing.T) {

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// State snapshotting
	if err := evm.TakeSnapshot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := evm.RevertSnapshot(); err != nil {
			t.Fatal(err)
		}
	})

	// Register node
	if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Mint & stake RPL
	rplAmount := eth.EthToWei(1000)
	if err := nodeutils.StakeRPL(rp, ownerAccount, nodeAccount, rplAmount); err != nil {
		t.Fatal(err)
	}

	// Get & set rewards claim interval
	rewardsClaimIntervalTime, err := protocol.GetRewardsClaimIntervalTime(rp, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := protocol.BootstrapRewardsClaimIntervalTime(rp, 0, ownerAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Check initial staking details
	if totalRplStake, err := node.GetTotalRPLStake(rp, nil); err != nil {
		t.Error(err)
	} else if totalRplStake.Cmp(rplAmount) != 0 {
		t.Errorf("Incorrect initial total RPL stake %s", totalRplStake.String())
	}
	if nodeRplStake, err := node.GetNodeRPLStake(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeRplStake.Cmp(rplAmount) != 0 {
		t.Errorf("Incorrect initial node RPL stake %s", nodeRplStake.String())
	}

	// Withdraw RPL
	if _, err := node.WithdrawRPL(rp, rplAmount, nodeAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

	// Check updated staking details
	if totalRplStake, err := node.GetTotalRPLStake(rp, nil); err != nil {
		t.Error(err)
	} else if totalRplStake.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect updated total RPL stake %s", totalRplStake.String())
	}
	if nodeRplStake, err := node.GetNodeRPLStake(rp, nodeAccount.Address, nil); err != nil {
		t.Error(err)
	} else if nodeRplStake.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Incorrect updated node RPL stake %s", nodeRplStake.String())
	}

	// Reset rewards claim interval
	if _, err := protocol.BootstrapRewardsClaimIntervalTime(rp, rewardsClaimIntervalTime, ownerAccount.GetTransactor()); err != nil {
		t.Fatal(err)
	}

}
