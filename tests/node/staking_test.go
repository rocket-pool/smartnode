package node

import (
    "math/big"
    "testing"

    "github.com/rocket-pool/rocketpool-go/node"
    "github.com/rocket-pool/rocketpool-go/tokens"
    "github.com/rocket-pool/rocketpool-go/utils/eth"

    "github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
    rplutils "github.com/rocket-pool/rocketpool-go/tests/testutils/tokens/rpl"
)


func TestStakeRPL(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register node
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Mint RPL
    rplAmount := eth.EthToWei(1000)
    if err := rplutils.MintRPL(rp, ownerAccount, nodeAccount, rplAmount); err != nil { t.Fatal(err) }

    // Approve RPL transfer for staking
    rocketNodeStakingAddress, err := rp.GetAddress("rocketNodeStaking")
    if err != nil { t.Fatal(err) }    
    if _, err := tokens.ApproveRPL(rp, *rocketNodeStakingAddress, rplAmount, nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

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
    if nodeRplStakedBlock, err := node.GetNodeRPLStakedBlock(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if nodeRplStakedBlock != 0 {
        t.Errorf("Incorrect initial node RPL staked block %d", nodeRplStakedBlock)
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
    } else if totalRplStake.Cmp(big.NewInt(0)) != 1 {
        t.Errorf("Incorrect updated total RPL stake 1 %s", totalRplStake.String())
    }
    if totalEffectiveRplStake, err := node.GetTotalEffectiveRPLStake(rp, nil); err != nil {
        t.Error(err)
    } else if totalEffectiveRplStake.Cmp(big.NewInt(0)) != 0 {
        t.Errorf("Incorrect updated total effective RPL stake 1 %s", totalEffectiveRplStake.String())
    }
    if nodeRplStake, err := node.GetNodeRPLStake(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if nodeRplStake.Cmp(big.NewInt(0)) != 1 {
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
    if nodeRplStakedBlock, err := node.GetNodeRPLStakedBlock(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if nodeRplStakedBlock == 0 {
        t.Errorf("Incorrect updated node RPL staked block 1 %d", nodeRplStakedBlock)
    }
    if nodeMinipoolLimit, err := node.GetNodeMinipoolLimit(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if nodeMinipoolLimit == 0 {
        t.Errorf("Incorrect updated node minipool limit 1 %d", nodeMinipoolLimit)
    }

    // Make node deposit to create minipool
    opts := nodeAccount.GetTransactor()
    opts.Value = eth.EthToWei(16)
    if _, err := node.Deposit(rp, 0, opts); err != nil { t.Fatal(err) }

    // Check updated staking details
    if totalEffectiveRplStake, err := node.GetTotalEffectiveRPLStake(rp, nil); err != nil {
        t.Error(err)
    } else if totalEffectiveRplStake.Cmp(big.NewInt(0)) != 1 {
        t.Errorf("Incorrect updated total effective RPL stake 2 %s", totalEffectiveRplStake.String())
    }
    if nodeEffectiveRplStake, err := node.GetNodeEffectiveRPLStake(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if nodeEffectiveRplStake.Cmp(big.NewInt(0)) != 1 {
        t.Errorf("Incorrect updated node effective RPL stake 2 %s", nodeEffectiveRplStake.String())
    }
    if nodeMinimumRplStake, err := node.GetNodeMinimumRPLStake(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if nodeMinimumRplStake.Cmp(big.NewInt(0)) != 1 {
        t.Errorf("Incorrect updated node minimum RPL stake 2 %s", nodeMinimumRplStake.String())
    }

}


func TestWithdrawRPL(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

}

