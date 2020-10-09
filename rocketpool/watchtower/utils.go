package watchtower

import (
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/rocket-pool/rocketpool-go/minipool"
    "github.com/rocket-pool/rocketpool-go/rocketpool"
    "github.com/rocket-pool/rocketpool-go/types"
    "golang.org/x/sync/errgroup"

    "github.com/rocket-pool/smartnode/shared/services/beacon"
)


// Get minipool validator statuses
func getMinipoolValidators(rp *rocketpool.RocketPool, bc beacon.Client, addresses []common.Address, callOpts *bind.CallOpts, validatorStatusOpts *beacon.ValidatorStatusOptions) (map[common.Address]beacon.ValidatorStatus, error) {

    // Data
    var wg errgroup.Group
    pubkeys := make([]types.ValidatorPubkey, len(addresses))

    // Get minipool validator pubkeys
    for mi, address := range addresses {
        mi, address := mi, address
        wg.Go(func() error {
            pubkey, err := minipool.GetMinipoolPubkey(rp, address, callOpts)
            if err == nil { pubkeys[mi] = pubkey }
            return err
        })
    }

    // Wait for data
    if err := wg.Wait(); err != nil {
        return map[common.Address]beacon.ValidatorStatus{}, err
    }

    // Get validator statuses
    statuses, err := bc.GetValidatorStatuses(pubkeys, validatorStatusOpts)
    if err != nil {
        return map[common.Address]beacon.ValidatorStatus{}, err
    }

    // Build validator map
    validators := make(map[common.Address]beacon.ValidatorStatus)
    for mi := 0; mi < len(addresses); mi++ {
        address := addresses[mi]
        pubkey := pubkeys[mi]
        status, ok := statuses[pubkey]
        if !ok { status = beacon.ValidatorStatus{} }
        validators[address] = status
    }

    // Return
    return validators, nil

}

