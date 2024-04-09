package rewards

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/eth"
	"github.com/rocket-pool/rocketpool-go/v2/types"
	"github.com/rocket-pool/smartnode/v2/rocketpool-daemon/common/state"
	"github.com/rocket-pool/smartnode/v2/shared"
	"github.com/rocket-pool/smartnode/v2/shared/keys"
	sharedtypes "github.com/rocket-pool/smartnode/v2/shared/types"
	"golang.org/x/sync/errgroup"
)

const (
	threadLimit int = 12
)

type RollingRecord struct {
	StartSlot         uint64                   `json:"startSlot"`
	LastDutiesSlot    uint64                   `json:"lastDutiesSlot"`
	ValidatorIndexMap map[string]*MinipoolInfo `json:"validatorIndexMap"`
	RewardsInterval   uint64                   `json:"rewardsInterval"`
	SmartnodeVersion  string                   `json:"smartnodeVersion,omitempty"`

	// Private fields
	bc                 beacon.IBeaconClient `json:"-"`
	beaconConfig       *beacon.Eth2Config   `json:"-"`
	genesisTime        time.Time            `json:"-"`
	logger             *slog.Logger         `json:"-"`
	intervalDutiesInfo *IntervalDutiesInfo  `json:"-"`

	// Constants for convenience
	one          *big.Int `json:"-"`
	validatorReq *big.Int `json:"-"`
}

// Create a new rolling record wrapper
func NewRollingRecord(logger *slog.Logger, bc beacon.IBeaconClient, startSlot uint64, beaconConfig *beacon.Eth2Config, rewardsInterval uint64) *RollingRecord {
	return &RollingRecord{
		StartSlot:         startSlot,
		LastDutiesSlot:    0,
		ValidatorIndexMap: map[string]*MinipoolInfo{},
		RewardsInterval:   rewardsInterval,
		SmartnodeVersion:  shared.RocketPoolVersion,

		bc:           bc,
		beaconConfig: beaconConfig,
		genesisTime:  time.Unix(int64(beaconConfig.GenesisTime), 0),
		logger:       logger,
		intervalDutiesInfo: &IntervalDutiesInfo{
			Slots: map[uint64]*SlotInfo{},
		},

		one:          eth.EthToWei(1),
		validatorReq: eth.EthToWei(32),
	}
}

// Load an existing record from serialized JSON data
func DeserializeRollingRecord(logger *slog.Logger, bc beacon.IBeaconClient, beaconConfig *beacon.Eth2Config, bytes []byte) (*RollingRecord, error) {
	record := &RollingRecord{
		bc:           bc,
		beaconConfig: beaconConfig,
		genesisTime:  time.Unix(int64(beaconConfig.GenesisTime), 0),
		logger:       logger,
		intervalDutiesInfo: &IntervalDutiesInfo{
			Slots: map[uint64]*SlotInfo{},
		},

		one:          eth.EthToWei(1),
		validatorReq: eth.EthToWei(32),
	}

	err := json.Unmarshal(bytes, &record)
	if err != nil {
		return nil, fmt.Errorf("error deserializing record: %w", err)
	}

	return record, nil
}

// Update the record to the requested slot, using the provided state as a reference.
// Requires the epoch *after* the requested slot to be finalized so it can accurately count attestations.
func (r *RollingRecord) UpdateToSlot(context context.Context, slot uint64, state *state.NetworkState) error {

	// Get the slot to start processing from
	startSlot := r.LastDutiesSlot + 1
	if r.LastDutiesSlot == 0 {
		startSlot = r.StartSlot
	}
	startEpoch := startSlot / r.beaconConfig.SlotsPerEpoch

	// Get the epoch for the state
	stateEpoch := slot / r.beaconConfig.SlotsPerEpoch

	//r.log.Printlnf("%s Updating rolling record from slot %d (epoch %d) to %d (epoch %d).", r.logPrefix, startSlot, startEpoch, slot, stateEpoch)
	//start := time.Now()

	// Update the validator indices and flag any cheating nodes
	r.updateValidatorIndices(state)

	// Process every epoch from the start to the current one
	for epoch := startEpoch; epoch <= stateEpoch; epoch++ {

		// Retrieve the duties for the epoch - this won't get duties higher than the given state
		err := r.getDutiesForEpoch(context, epoch, slot, state)
		if err != nil {
			return fmt.Errorf("error getting duties for epoch %d: %w", epoch, err)
		}

		// Process the epoch's attestation submissions
		err = r.processAttestationsInEpoch(context, epoch, state)
		if err != nil {
			return fmt.Errorf("error processing attestations in epoch %d: %w", epoch, err)
		}

	}

	// Process the epoch after the last one to check for late attestations / attestations of the last slot
	err := r.processAttestationsInEpoch(context, stateEpoch+1, state)
	if err != nil {
		return fmt.Errorf("error processing attestations in epoch %d: %w", stateEpoch+1, err)
	}

	// Clear the duties cache since it's not required anymore
	r.intervalDutiesInfo = &IntervalDutiesInfo{
		Slots: map[uint64]*SlotInfo{},
	}

	return nil
}

// Get the minipool scores, along with the cumulative total score and count - ignores minipools that belonged to cheaters
func (r *RollingRecord) GetScores(cheatingNodes map[common.Address]bool) ([]*MinipoolInfo, *big.Int, uint64) {
	// Create a slice of minipools with legal (non-cheater) scores
	minipoolInfos := make([]*MinipoolInfo, 0, len(r.ValidatorIndexMap))

	// TODO: return a new slice of minipool infos that ignores all cheaters
	totalScore := big.NewInt(0)
	totalCount := uint64(0)
	for _, mpInfo := range r.ValidatorIndexMap {

		// Ignore nodes that cheated
		if cheatingNodes[mpInfo.NodeAddress] {
			continue
		}

		totalScore.Add(totalScore, &mpInfo.AttestationScore.Int)
		totalCount += uint64(mpInfo.AttestationCount)
		minipoolInfos = append(minipoolInfos, mpInfo)
	}

	return minipoolInfos, totalScore, totalCount
}

// Serialize the current record into a byte array
func (r *RollingRecord) Serialize() ([]byte, error) {
	// Clone the record
	clone := &RollingRecord{
		StartSlot:         r.StartSlot,
		LastDutiesSlot:    r.LastDutiesSlot,
		RewardsInterval:   r.RewardsInterval,
		SmartnodeVersion:  r.SmartnodeVersion,
		ValidatorIndexMap: map[string]*MinipoolInfo{},
	}

	// Remove minipool perf records with zero attestations from the serialization
	for pubkey, mp := range r.ValidatorIndexMap {
		if mp.AttestationCount > 0 || len(mp.MissingAttestationSlots) > 0 {
			clone.ValidatorIndexMap[pubkey] = mp
		}
	}

	// Serialize as JSON
	bytes, err := json.Marshal(clone)
	if err != nil {
		return nil, fmt.Errorf("error serializing rolling record: %w", err)
	}

	return bytes, nil
}

// Update the validator index map with any new validators on Beacon
func (r *RollingRecord) updateValidatorIndices(state *state.NetworkState) {
	// NOTE: this has to go through every index each time in order to handle out-of-order validators
	// or invalid validators that got created on the testnet with broken deposits
	for i := 0; i < len(state.MinipoolDetails); i++ {
		mpd := state.MinipoolDetails[i]
		pubkey := mpd.Pubkey

		validator, exists := state.ValidatorDetails[pubkey]
		if !exists {
			// Hit a validator that doesn't exist on Beacon yet
			continue
		}

		_, exists = r.ValidatorIndexMap[validator.Index]
		if !exists && mpd.Status == types.MinipoolStatus_Staking {
			// Validator exists and is staking but it hasn't been recorded yet, add it to the map and update the latest index so we don't remap stuff we've already seen
			minipoolInfo := &MinipoolInfo{
				Address:                 mpd.MinipoolAddress,
				ValidatorPubkey:         mpd.Pubkey,
				ValidatorIndex:          validator.Index,
				NodeAddress:             mpd.NodeAddress,
				MissingAttestationSlots: map[uint64]bool{},
				AttestationScore:        sharedtypes.NewQuotedBigInt(0),
			}
			r.ValidatorIndexMap[validator.Index] = minipoolInfo
		}
	}
}

// Get the attestation duties for the given epoch, up to (and including) the provided end slot
func (r *RollingRecord) getDutiesForEpoch(context context.Context, epoch uint64, endSlot uint64, state *state.NetworkState) error {
	lastSlotInEpoch := (epoch+1)*r.beaconConfig.SlotsPerEpoch - 1

	if r.LastDutiesSlot >= lastSlotInEpoch {
		// Already collected the duties for this epoch
		r.logger.Debug("All duties were already collected, skipping...", slog.Uint64(keys.EpochKey, epoch))
		return nil
	}

	// Get the attestation committees for the epoch
	committees, err := r.bc.GetCommitteesForEpoch(context, &epoch)
	if err != nil {
		return fmt.Errorf("error getting committees for epoch %d: %w", epoch, err)
	}
	defer committees.Release()

	// Crawl the committees
	for idx := 0; idx < committees.Count(); idx++ {
		slotIndex := committees.Slot(idx)
		if slotIndex < r.StartSlot || slotIndex > endSlot {
			// Ignore slots that are out of bounds
			continue
		}
		if slotIndex <= r.LastDutiesSlot {
			// Ignore slots that have already been processed
			continue
		}
		blockTime := r.genesisTime.Add(time.Second * time.Duration(r.beaconConfig.SecondsPerSlot*slotIndex))
		committeeIndex := committees.Index(idx)

		// Check if there are any RP validators in this committee
		rpValidators := map[int]*MinipoolInfo{}
		for position, validator := range committees.Validators(idx) {
			mpInfo, exists := r.ValidatorIndexMap[validator]
			if !exists {
				// This isn't an RP validator, so ignore it
				continue
			}

			// Check if this minipool was opted into the SP for this block
			nodeDetails := state.NodeDetailsByAddress[mpInfo.NodeAddress]
			isOptedIn := nodeDetails.SmoothingPoolRegistrationState
			spRegistrationTime := time.Unix(nodeDetails.SmoothingPoolRegistrationChanged.Int64(), 0)
			if (isOptedIn && blockTime.Sub(spRegistrationTime) < 0) || // If this block occurred before the node opted in, ignore it
				(!isOptedIn && spRegistrationTime.Sub(blockTime) < 0) { // If this block occurred after the node opted out, ignore it
				continue
			}

			// Check if this minipool was in the `staking` state during this time
			mpd := state.MinipoolDetailsByAddress[mpInfo.Address]
			statusChangeTime := time.Unix(mpd.StatusTime.Int64(), 0)
			if mpd.Status != types.MinipoolStatus_Staking || blockTime.Sub(statusChangeTime) < 0 {
				continue
			}

			// This was a legal RP validator opted into the SP during this slot so add it
			rpValidators[position] = mpInfo
			mpInfo.MissingAttestationSlots[slotIndex] = true
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

	// Set the last slot duties were collected for - the minimum of the last slot in the epoch and the target state slot
	r.LastDutiesSlot = lastSlotInEpoch
	if endSlot < lastSlotInEpoch {
		r.LastDutiesSlot = endSlot
	}
	return nil
}

// Process the attestations proposed within the given epoch against the existing record, using
// the provided state for EL <-> CL mapping
func (r *RollingRecord) processAttestationsInEpoch(context context.Context, epoch uint64, state *state.NetworkState) error {
	slotsPerEpoch := r.beaconConfig.SlotsPerEpoch
	var wg errgroup.Group
	wg.SetLimit(threadLimit)
	attestationsPerSlot := make([][]beacon.AttestationInfo, r.beaconConfig.SlotsPerEpoch)

	// Get the attestation records for this epoch
	for i := uint64(0); i < slotsPerEpoch; i++ {
		i := i
		slot := epoch*slotsPerEpoch + i
		wg.Go(func() error {
			attestations, found, err := r.bc.GetAttestations(context, fmt.Sprint(slot))
			if err != nil {
				return fmt.Errorf("error getting attestations for slot %d: %w", slot, err)
			}
			if found {
				attestationsPerSlot[i] = attestations
			} else {
				attestationsPerSlot[i] = []beacon.AttestationInfo{}
			}

			return nil
		})
	}

	err := wg.Wait()
	if err != nil {
		return fmt.Errorf("error getting attestation records for epoch %d: %w", epoch, err)
	}

	// Process all of the slots in the epoch
	for i, attestations := range attestationsPerSlot {
		if len(attestations) > 0 {
			// Process these attestations
			slot := epoch*slotsPerEpoch + uint64(i)
			r.processAttestationsInSlot(slot, attestations, state)
		}
	}

	return nil
}

// Process all of the attestations for a given slot
func (r *RollingRecord) processAttestationsInSlot(inclusionSlot uint64, attestations []beacon.AttestationInfo, state *state.NetworkState) {
	// Go through the attestations for the block
	for _, attestation := range attestations {

		// Get the RP committees for this attestation's slot and index
		slotInfo, exists := r.intervalDutiesInfo.Slots[attestation.SlotIndex]
		if exists && inclusionSlot-attestation.SlotIndex <= r.beaconConfig.SlotsPerEpoch { // Ignore attestations delayed by more than 32 slots
			rpCommittee, exists := slotInfo.Committees[attestation.CommitteeIndex]
			if exists {
				blockTime := r.genesisTime.Add(time.Second * time.Duration(r.beaconConfig.SecondsPerSlot*attestation.SlotIndex))

				// Check if each RP validator attested successfully
				for position, validator := range rpCommittee.Positions {
					if attestation.AggregationBits.BitAt(uint64(position)) {
						// This was seen, so remove it from the missing attestations
						delete(rpCommittee.Positions, position)
						if len(rpCommittee.Positions) == 0 {
							delete(slotInfo.Committees, attestation.CommitteeIndex)
						}
						if len(slotInfo.Committees) == 0 {
							delete(r.intervalDutiesInfo.Slots, attestation.SlotIndex)
						}
						delete(validator.MissingAttestationSlots, attestation.SlotIndex)

						// Get the pseudoscore for this attestation
						details := state.MinipoolDetailsByAddress[validator.Address]
						bond, fee := getMinipoolBondAndNodeFee(details, blockTime)
						minipoolScore := big.NewInt(0).Sub(r.one, fee)   // 1 - fee
						minipoolScore.Mul(minipoolScore, bond)           // Multiply by bond
						minipoolScore.Div(minipoolScore, r.validatorReq) // Divide by 32 to get the bond as a fraction of a total validator
						minipoolScore.Add(minipoolScore, fee)            // Total = fee + (bond/32)(1 - fee)

						// Add it to the minipool's score
						validator.AttestationScore.Add(&validator.AttestationScore.Int, minipoolScore)
						validator.AttestationCount++
					}
				}
			}
		}
	}
}
