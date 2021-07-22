package rewards

import (
    "context"
    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/rewards"
    "github.com/rocket-pool/rocketpool-go/settings/protocol"
    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
    minipoolutils "github.com/rocket-pool/rocketpool-go/tests/testutils/minipool"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/rocket-pool/rocketpool-go/utils/eth"
    "math/big"
    "testing"
)


func TestNodeRewards(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Constants
    oneDay := 24 * 60 * 60
    rewardInterval := oneDay

    // Register node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Set network parameters
    if _, err := protocol.BootstrapRewardsClaimIntervalTime(rp, uint64(rewardInterval), ownerAccount.GetTransactor()); err != nil { t.Fatal(err) }

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

    // Increase time until node claims are possible
    if err := evm.IncreaseTime(rewardInterval); err != nil { t.Fatal(err) }

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
    } else if _, err := protocol.BootstrapInflationStartTime(rp, header.Time + uint64(oneDay), ownerAccount.GetTransactor()); err != nil {
        t.Fatal(err)
    }

    // Increase time until rewards are available
    if err := evm.IncreaseTime(oneDay + oneDay); err != nil {
        t.Fatal(err)
    }

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

