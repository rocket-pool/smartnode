package test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/prysmaticlabs/go-bitfield"
	"github.com/rocket-pool/rocketpool-go/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/state"
)

type epoch uint64
type slot uint64
type validatorIndex string
type validatorIndexToCommitteeIndexMap map[validatorIndex]uint
type criticalDutiesSlotMap map[validatorIndex]map[slot]interface{}

func (v *validatorIndexToCommitteeIndexMap) set(vI validatorIndex, i uint) {
	if *v == nil {
		*v = make(validatorIndexToCommitteeIndexMap)
	}
	(*v)[vI] = i
}

type missedDutiesMap map[epoch]map[slot][]validatorIndex

func (missedDuties *missedDutiesMap) add(s slot, validator validatorIndex) {
	if *missedDuties == nil {
		*missedDuties = make(missedDutiesMap)
	}
	e := epoch(s / 32)
	_, ok := (*missedDuties)[e]

	if !ok {
		(*missedDuties)[e] = make(map[slot][]validatorIndex)
	}
	_, ok = (*missedDuties)[e][s]
	if !ok {
		(*missedDuties)[e][s] = make([]validatorIndex, 0)
	}
	(*missedDuties)[e][s] = append((*missedDuties)[e][s], validator)
}

func (missedDuties *missedDutiesMap) getCount(s slot) uint {
	e := epoch(s / 32)
	if _, ok := (*missedDuties)[e]; !ok {
		return 0
	}
	if _, ok := (*missedDuties)[e][s]; !ok {
		return 0
	}
	return uint(len((*missedDuties)[e][s]))
}

type missedEpochsMap map[validatorIndex]map[epoch]interface{}

func (missedEpochs *missedEpochsMap) set(v validatorIndex, s slot) {
	e := epoch(s / 32)
	if *missedEpochs == nil {
		*missedEpochs = make(missedEpochsMap)
	}
	_, ok := (*missedEpochs)[v]
	if !ok {
		(*missedEpochs)[v] = make(map[epoch]interface{})
	}
	(*missedEpochs)[v][e] = struct{}{}
}

func (missedEpochs *missedEpochsMap) validatorMissedEpoch(v validatorIndex, e epoch) bool {
	if _, ok := (*missedEpochs)[v]; !ok {
		return false
	}
	_, ok := (*missedEpochs)[v][e]
	return ok
}

type MockBeaconClient struct {
	state *state.NetworkState

	t      *testing.T
	blocks map[string]beacon.BeaconBlock

	// A map of epoch -> slot -> validator indices for missed duties
	missedDuties missedDutiesMap

	// A map of validator -> epoch -> {}
	// that tracks which epochs a validator has missed duties in
	missedEpochs missedEpochsMap

	// Count of validators
	validatorCount uint

	// A map of validator index -> order in the list
	validatorIndices validatorIndexToCommitteeIndexMap

	// A map of validator index to pubkey
	validatorPubkeys map[validatorIndex]types.ValidatorPubkey

	// A map of validator index to critical duties slots
	criticalDutiesSlots criticalDutiesSlotMap
}

func (m *MockBeaconClient) SetState(state *state.NetworkState) {
	m.state = state
	if m.validatorPubkeys == nil {
		m.validatorPubkeys = make(map[validatorIndex]types.ValidatorPubkey)
	}
	for _, v := range state.ValidatorDetails {
		if _, ok := m.validatorPubkeys[validatorIndex(v.Index)]; ok {
			m.t.Fatalf("Validator %s already set", v.Index)
		}
		m.validatorPubkeys[validatorIndex(v.Index)] = v.Pubkey
	}
}

type mockBeaconCommitteeSlot struct {
	validators []string
}

type MockBeaconCommittees struct {
	slots []mockBeaconCommitteeSlot
	epoch epoch
}

func NewMockBeaconClient(t *testing.T) *MockBeaconClient {
	return &MockBeaconClient{t: t}
}

func (bc *MockBeaconClient) GetBeaconBlock(slot string) (beacon.BeaconBlock, bool, error) {
	if block, ok := bc.blocks[slot]; ok {
		return block, true, nil
	}
	return beacon.BeaconBlock{}, false, nil
}

func (bc *MockBeaconClient) SetBeaconBlock(slot string, block beacon.BeaconBlock) {
	if bc.blocks == nil {
		bc.blocks = make(map[string]beacon.BeaconBlock)
	}
	bc.blocks[slot] = block
}

func (bc *MockBeaconClient) SetCriticalDutiesSlots(criticalDutiesSlots *state.CriticalDutiesSlots) {
	if bc.criticalDutiesSlots == nil {
		bc.criticalDutiesSlots = make(criticalDutiesSlotMap)
	}
	for _validator, slots := range criticalDutiesSlots.CriticalDuties {
		validator := validatorIndex(_validator)
		if bc.criticalDutiesSlots[validator] == nil {
			bc.criticalDutiesSlots[validator] = make(map[slot]interface{})
		}
		for _, _slot := range slots {
			s := slot(_slot)
			bc.criticalDutiesSlots[validator][s] = struct{}{}
		}
	}
}

func (bc *MockBeaconClient) isValidatorActive(validator validatorIndex, e epoch) (bool, error) {
	// Get the pubkey
	validatorPubkey, ok := bc.validatorPubkeys[validator]
	if !ok {
		return false, fmt.Errorf("validator %s not found", validator)
	}
	validatorDetails, ok := bc.state.ValidatorDetails[validatorPubkey]
	if !ok {
		return false, fmt.Errorf("validator %s not found", validatorPubkey)
	}
	// Validators are assigned duties in the epoch they are activated
	// but not in the epoch they exit
	return validatorDetails.ActivationEpoch <= uint64(e) && (validatorDetails.ExitEpoch == 0 || uint64(e) < validatorDetails.ExitEpoch), nil
}

func (bc *MockBeaconClient) GetCommitteesForEpoch(_epoch *uint64) (beacon.Committees, error) {

	out := &MockBeaconCommittees{}
	out.epoch = epoch(*_epoch)

	// First find validators that must be assigned to specific slots
	var missedDutiesValidators map[slot][]validatorIndex
	missedDutiesValidators = bc.missedDuties[out.epoch]

	// Keep track of validators that have been assigned to a slot
	assignedValidators := make(map[string]interface{})

	out.slots = make([]mockBeaconCommitteeSlot, 32)
	for s := out.epoch * 32; s < out.epoch*32+32; s++ {
		idx := s - out.epoch*32
		out.slots[idx].validators = make([]string, 0, len(bc.validatorIndices)/32)

		// Assign validators that missed duties for this slot
		for _, validator := range missedDutiesValidators[slot(s)] {
			out.slots[idx].validators = append(out.slots[idx].validators, string(validator))
		}
		for _, validator := range out.slots[idx].validators {
			assignedValidators[validator] = struct{}{}
		}
	}

	// Assign the remaining validators based on total order / critical duties
	for validator, _ := range bc.validatorIndices {
		if _, ok := assignedValidators[string(validator)]; ok {
			continue
		}

		// If the validator was not active, skip it
		active, err := bc.isValidatorActive(validator, out.epoch)
		if err != nil {
			return nil, err
		}
		if !active {
			continue
		}

		// If the validator has critical duties for this slot, assign it
		if _, ok := bc.criticalDutiesSlots[validator]; ok {
			assigned := false
			for s, _ := range bc.criticalDutiesSlots[validator] {
				if bc.state.BeaconConfig.SlotToEpoch(uint64(s)) == uint64(out.epoch) {
					idx := s % 32
					out.slots[idx].validators = append(out.slots[idx].validators, string(validator))
					assigned = true
					break
				}
			}
			if assigned {
				continue
			}
		}

		// The validator was not assigned to a slot, neither by missing duties nor critical duties
		// Assign it to a pseudorandom slot
		idx := validator.Mod32()
		out.slots[idx].validators = append(out.slots[idx].validators, string(validator))
	}

	return out, nil
}

func (v validatorIndex) Mod32() uint {
	vInt, err := strconv.ParseUint(string(v), 10, 64)
	if err != nil {
		panic(err)
	}
	return uint(vInt % 32)
}

func (bc *MockBeaconClient) GetAttestations(_slot string) ([]beacon.AttestationInfo, bool, error) {

	slotNative, err := strconv.ParseUint(_slot, 10, 64)
	if err != nil {
		bc.t.Fatalf("Invalid slot: %s", _slot)
	}
	s := slot(slotNative)

	// Report attestations for the previous slot
	s -= 16

	// Get the epoch of the previous slot
	e := epoch(s / 32)

	// The length of the bitlist is the number of validators that missed duties
	// for the slot, plus the number of validators whose mod 32 is the same as the slot,
	// unless that validator has missed duties in the same epoch.
	//
	// However, a validator can be both in the set of validators that missed duties for the slot
	// and the set of validators whose mod 32 is the same as the slot, so we have to be careful
	// to not double count them.
	slotMod32 := s % 32
	var bitlistLength uint
	// Add the number of validators that missed duties for the slot
	bitlistLength = bc.missedDuties.getCount(s)

	for index, _ := range bc.validatorIndices {
		// Don't count validators that are have misses anywhere in this epoch
		if bc.missedEpochs.validatorMissedEpoch(index, e) {
			// This validator either missed this slot and was already counted,
			// or missed a different slot in the same epoch, and shouldn't be counted
			continue
		}

		active, err := bc.isValidatorActive(index, e)
		if err != nil {
			bc.t.Fatalf("Error checking if validator %s is active: %v", index, err)
		}
		if !active {
			continue
		}

		// Don't count validators with critical duties in this epoch unless the duty is in slot s
		if duties, ok := bc.criticalDutiesSlots[index]; ok {
			// The validator has some critical duties
			if _, ok := duties[s]; ok {
				// The duty is in slot s, so count it
				bitlistLength++
			} else {
				// Check if any duties are in the same epoch
				foundDuty := false
				for criticalDutySlot, _ := range duties {
					if bc.state.BeaconConfig.SlotToEpoch(uint64(criticalDutySlot)) == uint64(e) {
						foundDuty = true
						break
					}
				}
				if foundDuty {
					continue
				}
			}
		}

		// This validator was assigned to this slot and did not miss duties.
		validatorIndexMod32 := index.Mod32()
		if validatorIndexMod32 == uint(slotMod32) {
			bitlistLength++
		}
	}

	bl := bitfield.NewBitlist(uint64(bitlistLength))
	// Include all validators
	bl = bl.Not()
	// Exclude validators that need to miss duties on the previous slot
	if _, ok := bc.missedDuties[e]; ok {
		if _, ok := bc.missedDuties[e][s]; ok {
			numMissed := len(bc.missedDuties[e][s])
			for i := 0; i < numMissed; i++ {
				bl.SetBitAt(uint64(i), false)
			}
		}
	}
	out := []beacon.AttestationInfo{
		{
			AggregationBits: bl,
			SlotIndex:       uint64(s),
			CommitteeIndex:  0,
		},
	}
	return out, true, nil
}

// Count returns the number of committees in the response
func (mbc *MockBeaconCommittees) Count() int {
	return len(mbc.slots)
}

// Index returns the index of the committee at the provided offset
func (mbc *MockBeaconCommittees) Index(index int) uint64 {
	return 0
}

// Slot returns the slot of the committee at the provided offset
func (mbc *MockBeaconCommittees) Slot(index int) uint64 {
	return uint64(mbc.epoch)*32 + uint64(index)
}

// Validators returns the list of validators of the committee at
// the provided offset
func (mbc *MockBeaconCommittees) Validators(index int) []string {
	return mbc.slots[index].validators
}

// Release is a no-op
func (mbc *MockBeaconCommittees) Release() {
}

// SetMinipoolPerformance notes the minipool's performance
// to be mocked in the response to GetAttestations
func (bc *MockBeaconClient) SetMinipoolPerformance(index string, missedSlots []uint64) {

	// For each missed slot, add it to the inner map of slot to validator indices
	for _, s := range missedSlots {
		bc.missedDuties.add(slot(s), validatorIndex(index))

		// Add to missedEpochs
		bc.missedEpochs.set(validatorIndex(index), slot(s))
	}

	// A map of true validator index -> committee index
	if _, ok := bc.validatorIndices[validatorIndex(index)]; ok {
		bc.t.Fatalf("Validator %s already set", index)
	}
	bc.validatorIndices.set(validatorIndex(index), bc.validatorCount)
	bc.validatorCount++
}
