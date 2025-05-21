package rp

import (
	"bytes"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/smartnode/bindings/minipool"
	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/types"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
)

// Settings
const MinipoolPubkeyBatchSize = 50

// Get minipool validator statuses
func GetMinipoolValidators(rp *rocketpool.RocketPool, bc beacon.Client, addresses []common.Address, callOpts *bind.CallOpts, validatorStatusOpts *beacon.ValidatorStatusOptions) (map[common.Address]beacon.ValidatorStatus, error) {

	// Load minipool validator pubkeys in batches
	pubkeys := make([]types.ValidatorPubkey, len(addresses))
	for bsi := 0; bsi < len(addresses); bsi += MinipoolPubkeyBatchSize {

		// Get batch start & end index
		msi := bsi
		mei := min(bsi+MinipoolPubkeyBatchSize, len(addresses))

		// Load details
		var wg errgroup.Group
		for mi := msi; mi < mei; mi++ {
			mi := mi
			wg.Go(func() error {
				address := addresses[mi]
				pubkey, err := minipool.GetMinipoolPubkey(rp, address, callOpts)
				if err == nil {
					pubkeys[mi] = pubkey
				}
				return err
			})
		}
		if err := wg.Wait(); err != nil {
			return map[common.Address]beacon.ValidatorStatus{}, err
		}

	}

	// Filter out null and duplicate pubkeys
	filteredPubkeys := []types.ValidatorPubkey{}
	for _, pubkey := range pubkeys {
		if bytes.Equal(pubkey.Bytes(), types.ValidatorPubkey{}.Bytes()) {
			continue
		}
		isDuplicate := false
		for _, pk := range filteredPubkeys {
			if bytes.Equal(pubkey.Bytes(), pk.Bytes()) {
				isDuplicate = true
				break
			}
		}
		if isDuplicate {
			continue
		}
		filteredPubkeys = append(filteredPubkeys, pubkey)
	}

	// Get validator statuses
	statuses, err := bc.GetValidatorStatuses(filteredPubkeys, validatorStatusOpts)
	if err != nil {
		return map[common.Address]beacon.ValidatorStatus{}, err
	}

	// Build validator map
	validators := make(map[common.Address]beacon.ValidatorStatus)
	for mi := 0; mi < len(addresses); mi++ {
		address := addresses[mi]
		pubkey := pubkeys[mi]
		status, ok := statuses[pubkey]
		if !ok {
			status = beacon.ValidatorStatus{}
		}
		validators[address] = status
	}

	// Return
	return validators, nil

}
