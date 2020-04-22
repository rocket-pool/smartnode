package minipools

import (
    "encoding/hex"
    "errors"
    "fmt"

    "github.com/ethereum/go-ethereum/common"

    "github.com/rocket-pool/smartnode/shared/services/beacon"
    "github.com/rocket-pool/smartnode/shared/services/rocketpool/minipool"
    hexutil "github.com/rocket-pool/smartnode/shared/utils/hex"
)


/**
 * Check staking minipools for withdrawal
 */
func (p *MinipoolsProcess) checkStakingMinipools(minipoolAddresses []*common.Address) {

    // Check address count
    if len(minipoolAddresses) == 0 { return }

    // Log
    p.p.Log.Println("Checking staking minipools for withdrawal...")

    // Get current beacon head
    head, err := p.p.Beacon.GetBeaconHead()
    if err != nil {
        p.p.Log.Println(errors.New("Error retrieving current beacon head: " + err.Error()))
        return
    }

    // Check minipools
    for _, minipoolAddress := range minipoolAddresses {
        go p.checkStakingMinipool(minipoolAddress, head)
    }

}


/**
 * Check a staking minipool for withdrawal
 */
func (p *MinipoolsProcess) checkStakingMinipool(minipoolAddress *common.Address, head *beacon.BeaconHead) {

    // Log
    p.p.Log.Println(fmt.Sprintf("Checking minipool %s for withdrawal at epoch %d...", minipoolAddress.Hex(), head.Epoch))

    // Get minipool status
    status, err := minipool.GetStatus(p.p.CM, minipoolAddress)
    if err != nil {
        p.p.Log.Println(errors.New(fmt.Sprintf("Error retrieving minipool %s status: " + err.Error(), minipoolAddress.Hex())))
        return
    }

    // Get & check validator status; get minipool exit epoch
    validator, err := p.p.Beacon.GetValidatorStatus(hexutil.AddPrefix(hex.EncodeToString(status.ValidatorPubkey)))
    if err != nil {
        p.p.Log.Println(errors.New(fmt.Sprintf("Error retrieving minipool %s validator status: " + err.Error(), minipoolAddress.Hex())))
        return
    } else if !validator.Exists {
        p.p.Log.Println(fmt.Sprintf("Minipool %s validator does not yet exist on beacon chain...", minipoolAddress.Hex()))
        return
    }
    exitEpoch := validator.ActivationEpoch + status.StakingDuration.Uint64()

    // Check exit epoch
    if head.Epoch < exitEpoch {
        p.p.Log.Println(fmt.Sprintf("Minipool %s is not ready to withdraw until epoch %d...", minipoolAddress.Hex(), exitEpoch))
        return
    } else {
        p.p.Log.Println(fmt.Sprintf("Minipool %s is ready to withdraw, since epoch %d...", minipoolAddress.Hex(), exitEpoch))
    }

    // Withdraw minipool
    p.withdrawStakingMinipool(minipoolAddress)

}


/**
 * Withdraw a minipool
 */
func (p *MinipoolsProcess) withdrawStakingMinipool(minipoolAddress *common.Address) {

    // Log
    p.p.Log.Println("Minipool withdrawal process not yet implemented...")

}

