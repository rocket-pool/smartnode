package minipool

import (
	"bytes"
    "github.com/rocket-pool/rocketpool-go/types"
    "testing"

	"github.com/rocket-pool/rocketpool-go/minipool"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/utils/eth"

	"github.com/rocket-pool/rocketpool-go/tests/testutils/evm"
	minipoolutils "github.com/rocket-pool/rocketpool-go/tests/testutils/minipool"
	nodeutils "github.com/rocket-pool/rocketpool-go/tests/testutils/node"
	"github.com/rocket-pool/rocketpool-go/tests/testutils/validator"
)


func TestMinipoolDetails(t *testing.T) {

    // State snapshotting
    if err := evm.TakeSnapshot(); err != nil { t.Fatal(err) }
    t.Cleanup(func() { if err := evm.RevertSnapshot(); err != nil { t.Fatal(err) } })

    // Register nodes
    if _, err := node.RegisterNode(rp, "Australia/Brisbane", nodeAccount.GetTransactor()); err != nil { t.Fatal(err) }
    if err := nodeutils.RegisterTrustedNode(rp, ownerAccount, trustedNodeAccount); err != nil { t.Fatal(err) }

    // Get & check initial minipool details
    if minipools, err := minipool.GetMinipools(rp, nil); err != nil {
        t.Error(err)
    } else if len(minipools) != 0 {
        t.Error("Incorrect initial minipool count")
    }
    if nodeMinipools, err := minipool.GetNodeMinipools(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if len(nodeMinipools) != 0 {
        t.Error("Incorrect initial node minipool count")
    }
    if nodeMinipoolPubkeys, err := minipool.GetNodeValidatingMinipoolPubkeys(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if len(nodeMinipoolPubkeys) != 0 {
        t.Error("Incorrect initial node minipool pubkeys count")
    }

    // Minipool deposit/withdrawal amounts
    minipoolDepositAmount := eth.EthToWei(32)

    // Create & stake minipool
    mp, err := minipoolutils.CreateMinipool(rp, ownerAccount, nodeAccount, minipoolDepositAmount)
    if err != nil { t.Fatal(err) }
    if err := minipoolutils.StakeMinipool(rp, mp, nodeAccount); err != nil { t.Fatal(err) }

    // Mark minipool as withdrawable
    if _, err := minipool.SubmitMinipoolWithdrawable(rp, mp.Address, trustedNodeAccount.GetTransactor()); err != nil { t.Fatal(err) }

    // Get minipool validator pubkey
    validatorPubkey, err := validator.GetValidatorPubkey()
    if err != nil { t.Fatal(err) }

    // Get & check updated minipool details
    if minipools, err := minipool.GetMinipools(rp, nil); err != nil {
        t.Error(err)
    } else if len(minipools) != 1 {
        t.Error("Incorrect updated minipool count")
    } else {
        mpDetails := minipools[0]
        if !bytes.Equal(mpDetails.Address.Bytes(), mp.Address.Bytes()) {
            t.Errorf("Incorrect minipool address %s", mpDetails.Address.Hex())
        }
        if !mpDetails.Exists {
            t.Error("Incorrect minipool exists status")
        }
        if !bytes.Equal(mpDetails.Pubkey.Bytes(), validatorPubkey.Bytes()) {
            t.Errorf("Incorrect minipool validator pubkey %s", mpDetails.Pubkey.Hex())
        }
    }
    // Check status
    if status, err := mp.GetStatus(nil); err != nil {
        t.Error(err)
    } else {
        if status != types.Withdrawable {
            t.Error("Incorrect minipool withdrawable status")
        }
    }
    if nodeMinipools, err := minipool.GetNodeMinipools(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if len(nodeMinipools) != 1 {
        t.Error("Incorrect updated node minipool count")
    } else if !bytes.Equal(nodeMinipools[0].Address.Bytes(), mp.Address.Bytes()) {
        t.Errorf("Incorrect node minipool address %s", nodeMinipools[0].Address.Hex())
    }
    if nodeMinipoolPubkeys, err := minipool.GetNodeValidatingMinipoolPubkeys(rp, nodeAccount.Address, nil); err != nil {
        t.Error(err)
    } else if len(nodeMinipoolPubkeys) != 1 {
        t.Error("Incorrect updated node minipool pubkeys count")
    } else if !bytes.Equal(nodeMinipoolPubkeys[0].Bytes(), validatorPubkey.Bytes()) {
        t.Errorf("Incorrect node minipool pubkey %s", nodeMinipoolPubkeys[0].Hex())
    }

    // Get & check minipool address by pubkey
    if minipoolAddress, err := minipool.GetMinipoolByPubkey(rp, validatorPubkey, nil); err != nil {
        t.Error(err)
    } else if !bytes.Equal(minipoolAddress.Bytes(), mp.Address.Bytes()) {
        t.Errorf("Incorrect minipool address %s for pubkey %s", minipoolAddress.Hex(), validatorPubkey.Hex())
    }

}

