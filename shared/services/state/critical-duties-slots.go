package state

import (
	"github.com/rocket-pool/smartnode/shared/services/beacon"
)

type CriticalDutiesEpochs struct {
	// Map of epoch uint64 to a list of validator indices
	CriticalDuties map[uint64][]string
}

type CriticalDutiesSlots struct {
	// Map of validator index to a list of critical duties slots
	CriticalDuties map[string][]uint64
}

// Gets the critical duties slots for a given state as if it were the final state in a epochs epoch interval
func NewCriticalDutiesEpochs(epochs uint64, state *NetworkState) *CriticalDutiesEpochs {
	criticalDuties := &CriticalDutiesEpochs{
		CriticalDuties: make(map[uint64][]string),
	}

	endSlot := state.BeaconSlotNumber
	endEpoch := state.BeaconConfig.SlotToEpoch(endSlot)
	// Coerce endSlot to the last slot of the epoch
	endSlot = state.BeaconConfig.LastSlotOfEpoch(endEpoch)
	// Get the start epoch. Since the end epoch is the last inclusive epoch, we need to subtract 1 from the start epoch
	startEpoch := endEpoch - epochs - 1

	// Check for bond reductions first
	for _, minipool := range state.MinipoolDetails {
		lastReductionSlot := state.BeaconConfig.FirstSlotAtLeast(minipool.LastBondReductionTime.Int64())
		lastReductionEpoch := state.BeaconConfig.SlotToEpoch(lastReductionSlot)
		if lastReductionEpoch < startEpoch {
			continue
		}

		if lastReductionEpoch > endEpoch {
			continue
		}

		pubkey := minipool.Pubkey
		validatorIndex := state.ValidatorDetails[pubkey].Index
		criticalDuties.CriticalDuties[lastReductionEpoch] = append(criticalDuties.CriticalDuties[lastReductionEpoch], validatorIndex)
	}

	// Check for smoothing pool opt status changes next
	for _, node := range state.NodeDetails {
		lastOptStatusChange := state.BeaconConfig.FirstSlotAtLeast(node.SmoothingPoolRegistrationChanged.Int64())
		lastOptStatusChangeEpoch := state.BeaconConfig.SlotToEpoch(lastOptStatusChange)
		if lastOptStatusChangeEpoch < startEpoch {
			continue
		}

		if lastOptStatusChangeEpoch > endEpoch {
			continue
		}

		// Flag every minipool for this node as having a critical duty
		for _, minipool := range state.MinipoolDetailsByNode[node.NodeAddress] {
			pubkey := minipool.Pubkey
			validatorIndex := state.ValidatorDetails[pubkey].Index
			criticalDuties.CriticalDuties[lastOptStatusChangeEpoch] = append(criticalDuties.CriticalDuties[lastOptStatusChangeEpoch], validatorIndex)
		}
	}

	return criticalDuties
}

// For each validator in criticalDutiesEpochs, map the epochs to the slot the attestation duty assignment was for
func NewCriticalDutiesSlots(criticalDutiesEpochs *CriticalDutiesEpochs, bc beacon.Client) (*CriticalDutiesSlots, error) {
	criticalDuties := &CriticalDutiesSlots{
		CriticalDuties: make(map[string][]uint64),
	}

	for epoch, validatorIndices := range criticalDutiesEpochs.CriticalDuties {
		// Create a set of validator indices to query when iterating committees
		validatorIndicesSet := make(map[string]interface{})
		for _, validatorIndex := range validatorIndices {
			validatorIndicesSet[validatorIndex] = struct{}{}
		}

		// Get the beacon committee assignments for this epoch
		// Rebind e to avoid using a pointer to the accumulator.
		e := epoch
		committees, err := bc.GetCommitteesForEpoch(&e)
		if err != nil {
			return nil, err
		}

		// Iterate over the committees and check if the validator indices are in the set
		for i := 0; i < committees.Count(); i++ {
			validators := committees.Validators(i)
			for _, validator := range validators {
				if _, ok := validatorIndicesSet[validator]; ok {
					criticalDuties.CriticalDuties[validator] = append(criticalDuties.CriticalDuties[validator], committees.Slot(i))
				}
			}
		}
	}

	return criticalDuties, nil
}
