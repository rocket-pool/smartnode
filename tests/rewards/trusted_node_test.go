package rewards

import (
	"context"
	"math/big"
	"testing"

	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/settings/protocol"
	"github.com/rocket-pool/rocketpool-go/tokens"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
)


func TestTrustedNodeRewards(t *testing.T) {

    var secondsPerBlock uint64 = 12

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register node
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Set network parameters
    if _, err := protocol.BootstrapRewardsClaimIntervalTime(rp, 5 * secondsPerBlock, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if _, err := protocol.BootstrapInflationIntervalTime(rp, 5 * secondsPerBlock, ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get & check trusted node claims enabled status
    if claimsEnabled, err := rewards.GetTrustedNodeClaimsEnabled(rp, nil); err != nil {
        t.Error(err)
    } else if !claimsEnabled {
        t.Error("Incorrect trusted node claims enabled status")
    }

    // Get & check initial trusted node claim possible status
    if nodeClaimPossible, err := rewards.GetTrustedNodeClaimPossible(rp, trustedNodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if nodeClaimPossible {
        t.Error("Incorrect initial trusted node claim possible status")
    }

    // Mine blocks until trusted node claims are possible
    if err := evm.MineBlocks(5); err != nil { t.Fatal(err) }

    // Get & check updated trusted node claim possible status
    if nodeClaimPossible, err := rewards.GetTrustedNodeClaimPossible(rp, trustedNodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if !nodeClaimPossible {
        t.Error("Incorrect updated trusted node claim possible status")
    }

    // Get & check trusted node claim rewards percent
    if rewardsPerc, err := rewards.GetTrustedNodeClaimRewardsPerc(rp, trustedNodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if rewardsPerc != 1 {
        t.Errorf("Incorrect trusted node claim rewards perc %f", rewardsPerc)
    }

    // Get & check initial trusted node claim rewards amount
    if rewardsAmount, err := rewards.GetTrustedNodeClaimRewardsAmount(rp, trustedNodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if rewardsAmount.Cmp(big.NewInt(0)) != 0 {
        t.Errorf("Incorrect initial trusted node claim rewards amount %s", rewardsAmount.String())
    }

    // Start RPL inflation
    if header, err := rp.Client.HeaderByNumber(context.Background(), nil); err != nil {
        t.Fatal(err)
    } else if _, err := protocol.BootstrapInflationStartTime(rp, (header.Number.Uint64() + 2) * secondsPerBlock, ownerAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Mine blocks until rewards are available
    if err := evm.MineBlocks(10); err != nil { t.Fatal(err) }

    // Get & check updated trusted node claim rewards amount
    if rewardsAmount, err := rewards.GetTrustedNodeClaimRewardsAmount(rp, trustedNodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if rewardsAmount.Cmp(big.NewInt(0)) != 1 {
        t.Errorf("Incorrect updated trusted node claim rewards amount %s", rewardsAmount.String())
    }

    // Get & check initial node RPL balance
    if rplBalance, err := tokens.GetRPLBalance(rp, trustedNodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if rplBalance.Cmp(big.NewInt(0)) != 0 {
        t.Errorf("Incorrect initial node RPL balance %s", rplBalance.String())
    }

    // Claim node rewards
    if _, err := rewards.ClaimTrustedNodeRewards(rp, trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get & check updated node RPL balance
    if rplBalance, err := tokens.GetRPLBalance(rp, trustedNodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if rplBalance.Cmp(big.NewInt(0)) != 1 {
        t.Errorf("Incorrect updated node RPL balance %s", rplBalance.String())
    }

}

