package rewards

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/state"
)

type RollingRecord struct {
	StartSlot                 uint64
	CurrentSlot               uint64
	SuccessfulAttestations    uint64
	TotalAttestationScore     *big.Int
	MinipoolAttestationScores map[common.Address]*big.Int
	MinipoolAttestationCount  map[common.Address]uint64

	// Private fields
	updateLock        *sync.Mutex
	isRunning         bool
	wg                *sync.WaitGroup
	validatorIndexMap map[string]types.ValidatorPubkey
	latestMappedIndex int
}

// Create a new rolling record wrapper
func NewRollingRecord(startSlot uint64) *RollingRecord {
	return &RollingRecord{
		StartSlot:                 startSlot,
		CurrentSlot:               0,
		SuccessfulAttestations:    0,
		TotalAttestationScore:     big.NewInt(0),
		MinipoolAttestationScores: map[common.Address]*big.Int{},
		MinipoolAttestationCount:  map[common.Address]uint64{},
		updateLock:                &sync.Mutex{},
		isRunning:                 false,
		wg:                        nil,
		validatorIndexMap:         map[string]types.ValidatorPubkey{},
		latestMappedIndex:         -1,
	}
}

// Update the record to the provided state.
// Returns true if the update has been queued, or false if there is already an update in progress so this was ignored.
func (r *RollingRecord) UpdateToState(state *state.NetworkState) (bool, error) {

	// Return if there's an update in progress
	r.updateLock.Lock()
	if r.isRunning {
		r.updateLock.Unlock()
		return false, nil
	}

	// Set up a new goroutine
	r.isRunning = true
	r.wg = &sync.WaitGroup{}
	r.wg.Add(1)
	r.updateLock.Unlock()

	// Run the update logic
	go func() {
		r.updateValidatorIndices(state)
		r.removeCheaters(state)

		r.wg.Done()
	}()

	return true, nil
}

// Waits for the active update routine if it's running
func (r *RollingRecord) WaitForUpdate() {
	r.updateLock.Lock()
	isNil := (r.wg == nil)
	r.updateLock.Unlock()

	if !isNil {
		r.wg.Wait()
	}
}

// Update the validator index map with any new validators on Beacon
func (r *RollingRecord) updateValidatorIndices(state *state.NetworkState) {
	for i := r.latestMappedIndex + 1; i < len(state.MinipoolDetails); i++ {
		mpd := state.MinipoolDetails[i]
		pubkey := mpd.Pubkey

		validator, exists := state.ValidatorDetails[pubkey]
		if !exists {
			// Hit a validator that doesn't exist on Beacon yet
			return
		}

		// Validator exists, add it to the map and update the latest index so we don't remap stuff we've already seen
		r.validatorIndexMap[validator.Index] = pubkey
		r.latestMappedIndex = i
	}
}

// Detect and remove any cheaters
func (r *RollingRecord) removeCheaters(state *state.NetworkState) {
	three := big.NewInt(3)
	for _, nd := range state.NodeDetails {
		cheated := false
		for _, mpd := range state.MinipoolDetailsByNode[nd.NodeAddress] {
			if mpd.PenaltyCount.Cmp(three) >= 0 {
				// If any minipool has 3+ penalties, ban the entire node
				r.removeCheaterFromRecord(nd.NodeAddress, state)
				cheated = true
				break
			}
		}
		if cheated {
			continue
		}
	}
}

// Remove a node that's been flagged for cheating from the record
func (r *RollingRecord) removeCheaterFromRecord(nodeAddress common.Address, state *state.NetworkState) {
	for _, mpd := range state.MinipoolDetailsByNode[nodeAddress] {
		score, exists := r.MinipoolAttestationScores[mpd.MinipoolAddress]
		if !exists {
			// Ignore this MP if it wasn't in the record
			continue
		}

		r.TotalAttestationScore.Sub(r.TotalAttestationScore, score)
		r.SuccessfulAttestations -= r.MinipoolAttestationCount[mpd.MinipoolAddress]
		delete(r.MinipoolAttestationScores, mpd.MinipoolAddress)
		delete(r.MinipoolAttestationCount, mpd.MinipoolAddress)
	}
}
