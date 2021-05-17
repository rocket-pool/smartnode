package rewards

import (
	"context"
	"math/big"
	"testing"

	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	minipoolutils "github.com/rocket-pool/rocketpool-go/tests/testutils/minipool"
)


func TestNodeRewards(t *testing.T) {

    var secondsPerBlock uint64 = 12

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Set network parameters
    if _, err := protocol.BootstrapRewardsClaimIntervalTime(rp, 5 * secondsPerBlock, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if _, err := protocol.BootstrapInflationIntervalTime(rp, 5 * secondsPerBlock, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get & check node claims enabled status
    if claimsEnabled, err := rewards.GetNodeClaimsEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if !claimsEnabled {
        t.Error("Incorrect node claims enabled status")
    }

    // Get & check initial node claim possible status
    if nodeClaimPossible, err := rewards.GetNodeClaimPossible(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if nodeClaimPossible {
        t.Error("Incorrect initial node claim possible status")
    }

    // Mine blocks until node claims are possible
    if err := evm.MineBlocks(5); err != nil { t.Fatal(err) }
    if err := evm.IncreaseTime(5 * secondsPerBlock); err != nil { t.Fatal(err) }

    // Get & check updated node claim possible status
    if nodeClaimPossible, err := rewards.GetNodeClaimPossible(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if !nodeClaimPossible {
        t.Error("Incorrect updated node claim possible status")
    }

    // Get & check initial node claim rewards percent
    if rewardsPerc, err := rewards.GetNodeClaimRewardsPerc(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if rewardsPerc != 0 {
        t.Errorf("Incorrect initial node claim rewards perc %f", rewardsPerc)
    }

    // Stake RPL & create a minipool
    if _, err := minipoolutils.CreateMinipool(rp, ownerAccount, nodeAccount, eth.EthToWei(16)); err != nil { t.Fatal(err) }

    // Get & check updated node claim rewards percent
    if rewardsPerc, err := rewards.GetNodeClaimRewardsPerc(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if rewardsPerc != 1 {
        t.Errorf("Incorrect updated node claim rewards perc %f", rewardsPerc)
    }

    // Get & check initial node claim rewards amount
    if rewardsAmount, err := rewards.GetNodeClaimRewardsAmount(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if rewardsAmount.Cmp(big.NewInt(0)) != 0 {
        t.Errorf("Incorrect initial node claim rewards amount %s", rewardsAmount.String())
    }

    // Start RPL inflation
    if header, err := rp.Client.HeaderByNumber(context.Background(), nil); err != nil {
        t.Fatal(err)
    } else if _, err := protocol.BootstrapInflationStartTime(rp, (header.Number.Uint64() + 2) * secondsPerBlock, ownerAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Mine blocks until rewards are available
    if err := evm.MineBlocks(10); err != nil { t.Fatal(err) }
    if err := evm.IncreaseTime(10 * secondsPerBlock); err != nil { t.Fatal(err) }

    // Get & check updated node claim rewards amount
    if rewardsAmount, err := rewards.GetNodeClaimRewardsAmount(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if rewardsAmount.Cmp(big.NewInt(0)) != 1 {
        t.Errorf("Incorrect updated node claim rewards amount %s", rewardsAmount.String())
    }

    // Get & check initial node RPL balance
    if rplBalance, err := tokens.GetRPLBalance(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if rplBalance.Cmp(big.NewInt(0)) != 0 {
        t.Errorf("Incorrect initial node RPL balance %s", rplBalance.String())
    }

    // Claim node rewards
    if _, err := rewards.ClaimNodeRewards(rp, nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get & check updated node RPL balance
    if rplBalance, err := tokens.GetRPLBalance(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if rplBalance.Cmp(big.NewInt(0)) != 1 {
        t.Errorf("Incorrect updated node RPL balance %s", rplBalance.String())
    }

}

