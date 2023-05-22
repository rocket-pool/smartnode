package rewards

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	rpstate "github.com/rocket-pool/rocketpool-go/utils/state"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/state"
	"golang.org/x/sync/errgroup"
)

const (
	threadLimit int = 12
)

type RollingRecord struct {
	StartSlot         uint64
	LastProcessedSlot uint64
	UpdateError       error

	// Private fields
	bc                 beacon.Client
	beaconConfig       *beacon.Eth2Config
	updateLock         *sync.Mutex
	isRunning          bool
	wg                 *sync.WaitGroup
	validatorIndexMap  map[string]*MinipoolInfo
	latestMappedIndex  int
	intervalDutiesInfo *IntervalDutiesInfo
	cheatingNodes      map[common.Address]bool

	// Constants for convenience
	zero         *big.Int
	one          *big.Int
	validatorReq *big.Int
}

// Create a new rolling record wrapper
func NewRollingRecord(bc beacon.Client, startSlot uint64, beaconConfig *beacon.Eth2Config) *RollingRecord {
	return &RollingRecord{
		StartSlot:         startSlot,
		LastProcessedSlot: 0,

		bc:                bc,
		beaconConfig:      beaconConfig,
		updateLock:        &sync.Mutex{},
		isRunning:         false,
		wg:                nil,
		validatorIndexMap: map[string]*MinipoolInfo{},
		latestMappedIndex: -1,
		intervalDutiesInfo: &IntervalDutiesInfo{
			Slots: map[uint64]*SlotInfo{},
		},

		cheatingNodes: map[common.Address]bool{},
		zero:          big.NewInt(0),
		one:           eth.EthToWei(1),
		validatorReq:  eth.EthToWei(32),
	}
}

// Update the record to the provided state. If processAttestationsOnly is true, it won't collect attestation duties for the
// epoch represented by the provided state (but will for all prior epochs); it will just process its attestation submissions.
// NOTE: assumes the state is the last slot of its epoch, and that it has been finalized!
// Returns true if the update has been queued, or false if there is already an update in progress so this was ignored.
// Use RollingRecord.UpdateError to check if something went wrong with the previous update prior to running this again;
// otherwise calling this after an error from a previous iteration will just return that error.
func (r *RollingRecord) UpdateToState(state *state.NetworkState, processAttestationsOnly bool) (bool, error) {

	// Return if there's an update in progress
	r.updateLock.Lock()
	if r.isRunning {
		r.updateLock.Unlock()
		return false, nil
	}

	// Check the last error
	if r.UpdateError != nil {
		return false, r.UpdateError
	}

	// Set up a new goroutine
	r.isRunning = true
	r.wg = &sync.WaitGroup{}
	r.wg.Add(1)
	r.updateLock.Unlock()

	// Run the update logic
	go func() {
		// Update the validator indices and flag any cheating nodes
		r.updateValidatorIndices(state)
		r.flagCheaters(state)

		// Get the epoch to start processing from
		startEpoch := uint64(0)
		if r.LastProcessedSlot == 0 {
			// First starting up, so get the epoch from the start slot
			startEpoch = r.StartSlot / r.beaconConfig.SlotsPerEpoch
		} else {
			startEpoch = r.LastProcessedSlot / r.beaconConfig.SlotsPerEpoch
		}

		// Get the epoch for the state
		stateEpoch := state.BeaconSlotNumber / r.beaconConfig.SlotsPerEpoch

		// Process every epoch from the start to the current one
		for epoch := startEpoch; epoch <= stateEpoch; epoch++ {

			// Ignore duties from the final epoch if requested
			if epoch == stateEpoch && processAttestationsOnly {
				err := r.getDutiesForEpoch(epoch)
				if err != nil {
					r.UpdateError = fmt.Errorf("error getting duties for epoch %d: %w", epoch, err)
					return
				}
			}

			// Process the epoch's attestation submissions
			err := r.processAttestationsInEpoch(epoch, state)
			if err != nil {
				r.UpdateError = fmt.Errorf("error processing attestations in epoch %d: %w", epoch, err)
				return
			}

		}

		r.LastProcessedSlot = state.BeaconSlotNumber
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

// Get the minipool scores, along with the cumulative total score and count
func (r *RollingRecord) GetScores(state *state.NetworkState) (map[string]*MinipoolInfo, *big.Int, uint64) {

	totalScore := big.NewInt(0)
	totalCount := uint64(0)
	for _, mpInfo := range r.validatorIndexMap {
		// Ignore nodes that cheated
		if r.cheatingNodes[mpInfo.NodeAddress] {
			continue
		}

		totalScore.Add(totalScore, mpInfo.AttestationScore)
		totalCount += mpInfo.AttestationCount
	}

	return r.validatorIndexMap, totalScore, totalCount
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
		minipoolInfo := &MinipoolInfo{
			Address:                 mpd.MinipoolAddress,
			ValidatorPubkey:         mpd.Pubkey,
			ValidatorIndex:          validator.Index,
			NodeAddress:             mpd.NodeAddress,
			MissingAttestationSlots: map[uint64]bool{},
			AttestationScore:        big.NewInt(0),
		}
		r.validatorIndexMap[validator.Index] = minipoolInfo
		r.latestMappedIndex = i
	}
}

// Detect and flag any cheaters
func (r *RollingRecord) flagCheaters(state *state.NetworkState) {
	three := big.NewInt(3)
	for _, nd := range state.NodeDetails {
		for _, mpd := range state.MinipoolDetailsByNode[nd.NodeAddress] {
			if mpd.PenaltyCount.Cmp(three) >= 0 {
				// If any minipool has 3+ penalties, ban the entire node
				r.cheatingNodes[nd.NodeAddress] = true
				break
			}
		}
		if r.cheatingNodes[nd.NodeAddress] {
			continue
		}
	}
}

// Get the attestation duties for the given epoch
func (r *RollingRecord) getDutiesForEpoch(epoch uint64) error {

	// Get the attestation committees for the epoch
	committees, err := r.bc.GetCommitteesForEpoch(&epoch)
	if err != nil {
		return fmt.Errorf("error getting committees for epoch %d: %w", epoch, err)
	}
	defer committees.Release()

	// Crawl the committees
	lastSlot := epoch*r.beaconConfig.SlotsPerEpoch + (r.beaconConfig.SlotsPerEpoch - 1)
	for idx := 0; idx < committees.Count(); idx++ {
		slotIndex := committees.Slot(idx)
		if slotIndex < r.StartSlot || slotIndex > lastSlot {
			// Ignore slots that are out of bounds
			continue
		}
		committeeIndex := committees.Index(idx)

		// Check if there are any RP validators in this committee
		rpValidators := map[int]*MinipoolInfo{}
		for position, validator := range committees.Validators(idx) {
			mpInfo, exists := r.validatorIndexMap[validator]
			if exists && !r.cheatingNodes[mpInfo.NodeAddress] {
				// If it's an RP validator and the node isn't a cheater, this can be counted
				rpValidators[position] = mpInfo
				mpInfo.MissingAttestationSlots[slotIndex] = true
			}
		}

		// If there are some RP validators, add this committee to the map
		if len(rpValidators) > 0 {
			slotInfo, exists := r.intervalDutiesInfo.Slots[slotIndex]
			if !exists {
				slotInfo = &SlotInfo{
					Index:      slotIndex,
					Committees: map[uint64]*CommitteeInfo{},
				}
				r.intervalDutiesInfo.Slots[slotIndex] = slotInfo
			}
			slotInfo.Committees[committeeIndex] = &CommitteeInfo{
				Index:     committeeIndex,
				Positions: rpValidators,
			}
		}
	}

	return nil

}

// Process the attestations proposed within the given epoch
func (r *RollingRecord) processAttestationsInEpoch(epoch uint64, state *state.NetworkState) error {

	slotsPerEpoch := r.beaconConfig.SlotsPerEpoch
	var wg errgroup.Group
	wg.SetLimit(threadLimit)

	// Get the attestation records for this epoch
	for i := uint64(0); i < slotsPerEpoch; i++ {
		i := i
		slot := epoch*slotsPerEpoch + i
		wg.Go(func() error {
			attestations, found, err := r.bc.GetAttestations(fmt.Sprint(slot))
			if err != nil {
				return fmt.Errorf("error getting attestations for slot %d: %w", slot, err)
			}
			if found && len(attestations) > 0 {
				r.processAttestationsInSlot(attestations, state)
			}
			return nil
		})
	}

	err := wg.Wait()
	if err != nil {
		return fmt.Errorf("error getting attestation records for epoch %d: %w", epoch, err)
	}

	return nil

}

// Process all of the attestations for a given slot
func (r *RollingRecord) processAttestationsInSlot(attestations []beacon.AttestationInfo, state *state.NetworkState) {

	// Go through the attestations for the block
	for _, attestation := range attestations {

		// Get the RP committees for this attestation's slot and index
		slotInfo, exists := r.intervalDutiesInfo.Slots[attestation.SlotIndex]
		if exists {
			rpCommittee, exists := slotInfo.Committees[attestation.CommitteeIndex]
			if exists {
				blockTime := time.Unix(int64(r.beaconConfig.GenesisTime), 0).Add(time.Second * time.Duration(r.beaconConfig.SecondsPerSlot*attestation.SlotIndex))

				// Check if each RP validator attested successfully
				for position, validator := range rpCommittee.Positions {
					if attestation.AggregationBits.BitAt(uint64(position)) {
						// This was seen, so remove it from the missing attestations and add it to the completed ones
						delete(rpCommittee.Positions, position)
						if len(rpCommittee.Positions) == 0 {
							delete(slotInfo.Committees, attestation.CommitteeIndex)
						}
						if len(slotInfo.Committees) == 0 {
							delete(r.intervalDutiesInfo.Slots, attestation.SlotIndex)
						}
						validator.CompletedAttestations[attestation.SlotIndex] = true
						delete(validator.MissingAttestationSlots, attestation.SlotIndex)

						// Check if this minipool was opted into the SP for this block
						nodeDetails := state.NodeDetailsByAddress[validator.NodeAddress]
						isOptedIn := nodeDetails.SmoothingPoolRegistrationState
						statusChangeTime := time.Unix(nodeDetails.SmoothingPoolRegistrationChanged.Int64(), 0)
						if (isOptedIn && blockTime.Sub(statusChangeTime) < 0) || // If this block occurred before the node opted in, ignore it
							(!isOptedIn && statusChangeTime.Sub(blockTime) < 0) { // If this block occurred after the node opted out, ignore it
							// Not opted in
							continue
						}

						// Get the pseudoscore for this attestation
						details := state.MinipoolDetailsByAddress[validator.Address]
						bond, fee := r.getMinipoolBondAndNodeFee(details, blockTime)
						minipoolScore := big.NewInt(0).Sub(r.one, fee)   // 1 - fee
						minipoolScore.Mul(minipoolScore, bond)           // Multiply by bond
						minipoolScore.Div(minipoolScore, r.validatorReq) // Divide by 32 to get the bond as a fraction of a total validator
						minipoolScore.Add(minipoolScore, fee)            // Total = fee + (bond/32)(1 - fee)

						// Add it to the minipool's score
						validator.AttestationScore.Add(validator.AttestationScore, minipoolScore)
						validator.AttestationCount++
					}
				}
			}
		}
	}

}

// Get the bond and node fee of a minipool for the specified time
func (r *RollingRecord) getMinipoolBondAndNodeFee(details *rpstate.NativeMinipoolDetails, blockTime time.Time) (*big.Int, *big.Int) {
	currentBond := details.NodeDepositBalance
	currentFee := details.NodeFee
	previousBond := details.LastBondReductionPrevValue
	previousFee := details.LastBondReductionPrevNodeFee

	var reductionTimeBig *big.Int = details.LastBondReductionTime
	if reductionTimeBig.Cmp(r.zero) == 0 {
		// Never reduced
		return currentBond, currentFee
	} else {
		reductionTime := time.Unix(reductionTimeBig.Int64(), 0)
		if reductionTime.Sub(blockTime) > 0 {
			// This block occurred before the reduction
			if previousFee.Cmp(r.zero) == 0 {
				// Catch for minipools that were created before this call existed
				return previousBond, currentFee
			}
			return previousBond, previousFee
		}
	}

	return currentBond, currentFee
}
