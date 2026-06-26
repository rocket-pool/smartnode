// Package performance implements RPIP-73 target-vote performance verification.
//
// RPIP-73 measures attestation performance using the "target" timeliness flag
// from the Beacon State's previous_epoch_participation vector. The flag is
// defined to be set for validator v in epoch E iff some attestation by v with
// data.target.epoch == E was:
//
//  1. included in a block (which implies source-checkpoint matching), AND
//  2. voting for the correct target root, i.e. data.target.root equals the
//     canonical block root at the first slot of epoch E, AND
//  3. included within SLOTS_PER_EPOCH slots of data.slot.
//
// This package recomputes the flag by inspecting block attestations rather
// than downloading the full Beacon State SSZ. Per epoch under inspection the
// cost is roughly: one committees fetch + one block-header fetch (target
// root) + up to SLOTS_PER_EPOCH block fetches (inclusion window). This is
// orders of magnitude cheaper than fetching beacon states, and works against
// any standard Beacon API node (archival is still required for old slots).
package performance

import (
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"

	// "github.com/rocket-pool/smartnode/bindings/settings/protocol"
	// "github.com/rocket-pool/smartnode/bindings/utils/eth"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/types/api"
)

// Beacon participation flag indices, see the Ethereum consensus spec
// (Altair upgrade). previous_epoch_participation packs these as bit flags
// per validator in a single byte.
const (
	TimelySourceFlagIndex = 0
	TimelyTargetFlagIndex = 1
	TimelyHeadFlagIndex   = 2
)

// defaultPerformanceThresholdPct is the RPIP-73 initial pDAO
// performance_threshold value (94%). It is used while the on-chain
// rocketDAOProtocolSettingsPerformance contract is not yet deployed; once
// available, replace this with a live call to protocol.GetPerformanceThreshold.
const defaultPerformanceThresholdPct = 94.0

// farFutureEpoch is the spec's FAR_FUTURE_EPOCH sentinel (2^64-1).
const farFutureEpoch = ^uint64(0)

// maxTargetRootWalkback bounds the number of skipped slots we will walk back
// to find the canonical target block root for an epoch. In any healthy chain
// this is 0 (the boundary slot has a block) or a small handful.
const maxTargetRootWalkback = 64

// EpochPerformance is the per-epoch result of a target-vote check.
type EpochPerformance struct {
	Epoch        uint64 `json:"epoch"`
	TimelyTarget bool   `json:"timelyTarget"`
}

// PerformanceSummary aggregates the per-epoch results of a target-vote check
// over the inclusive range [StartEpoch, EndEpoch]. Epochs in which the
// validator was not assigned to any committee (i.e. not yet active or already
// exited) are counted in InactiveEpochs and excluded from PerformancePct's
// denominator.
type PerformanceSummary struct {
	ValidatorIndex  uint64   `json:"validatorIndex"`
	StartEpoch      uint64   `json:"startEpoch"`
	EndEpoch        uint64   `json:"endEpoch"`
	TotalEpochs     uint64   `json:"totalEpochs"`
	TimelyEpochs    uint64   `json:"timelyEpochs"`
	MissedEpochs    uint64   `json:"missedEpochs"`
	InactiveEpochs  uint64   `json:"inactiveEpochs"`
	PerformancePct  float64  `json:"performancePct"`
	MissedEpochList []uint64 `json:"missedEpochList"`
	TimelyEpochList []uint64 `json:"timelyEpochList"`
}

// PerformanceBeaconClient is the minimal beacon client surface used by the
// block-based target-vote engine.
type PerformanceBeaconClient interface {
	GetEth2Config() (beacon.Eth2Config, error)
	GetCommitteesForEpoch(epoch *uint64) (beacon.Committees, error)
	GetBeaconBlock(blockId string) (beacon.BeaconBlock, bool, error)
	GetBeaconBlockHeader(blockId string) (beacon.BeaconBlockHeader, bool, error)
	GetValidatorStatusByIndex(index string, opts *beacon.ValidatorStatusOptions) (beacon.ValidatorStatus, error)
}

// pubkeyBeaconClient is the beacon client surface needed to resolve a
// validator pubkey to a beacon-chain index in addition to the engine
// requirements.
type pubkeyBeaconClient interface {
	PerformanceBeaconClient
	GetValidatorIndex(pubkey rptypes.ValidatorPubkey) (string, error)
}

// VerifyPerformance is the end-to-end RPIP-73 target-vote verification flow
// shared by the minipool and megapool API endpoints. It resolves the
// validator's beacon-chain index from the supplied pubkey, runs
// CheckTargetPerformance, and packages the result alongside the pDAO
// performance_threshold for pass/fail reporting.
func VerifyPerformance(
	rp *rocketpool.RocketPool,
	bc pubkeyBeaconClient,
	pubkey rptypes.ValidatorPubkey,
	startEpoch uint64,
	endEpoch uint64,
) (*api.VerifyPerformanceResponse, error) {
	if pubkey == (rptypes.ValidatorPubkey{}) {
		return nil, fmt.Errorf("validator has no pubkey on-chain yet (not deposited?)")
	}

	cfg, err := bc.GetEth2Config()
	if err != nil {
		return nil, fmt.Errorf("error getting beacon config: %w", err)
	}

	indexStr, err := bc.GetValidatorIndex(pubkey)
	if err != nil {
		return nil, fmt.Errorf("error getting beacon-chain index for validator %s: %w", pubkey.Hex(), err)
	}
	validatorIndex, err := strconv.ParseUint(indexStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing validator index %q: %w", indexStr, err)
	}

	summary, err := CheckTargetPerformance(bc, cfg, validatorIndex, startEpoch, endEpoch)
	if err != nil {
		return nil, err
	}

	// The performance_threshold setting is scaled by 1e18 (1e18 = 100%).
	// thresholdWei, err := protocol.GetPerformanceThreshold(rp, nil)
	// if err != nil {
	// 	return nil, fmt.Errorf("error getting performance threshold: %w", err)
	// }
	// thresholdPct := eth.WeiToEth(thresholdWei) * 100
	//
	// TODO: The rocketDAOProtocolSettingsPerformance contract is not yet deployed.
	// Use the RPIP-73 initial value until it is.
	thresholdPct := defaultPerformanceThresholdPct

	return &api.VerifyPerformanceResponse{
		ValidatorPubkey:         pubkey,
		ValidatorIndex:          summary.ValidatorIndex,
		StartEpoch:              summary.StartEpoch,
		EndEpoch:                summary.EndEpoch,
		TotalEpochs:             summary.TotalEpochs,
		TimelyEpochs:            summary.TimelyEpochs,
		MissedEpochs:            summary.MissedEpochs,
		InactiveEpochs:          summary.InactiveEpochs,
		PerformancePct:          summary.PerformancePct,
		PerformanceThresholdPct: thresholdPct,
		PassesThreshold:         summary.PerformancePct >= thresholdPct,
		MissedEpochList:         summary.MissedEpochList,
		TimelyEpochList:         summary.TimelyEpochList,
	}, nil
}

// CheckTargetPerformance evaluates a single validator's target-vote
// performance over the inclusive epoch range [startEpoch, endEpoch] by
// reading the canonical target root, the validator's committee assignment,
// and the attestations in the inclusion window per epoch.
func CheckTargetPerformance(
	bc PerformanceBeaconClient,
	cfg beacon.Eth2Config,
	validatorIndex uint64,
	startEpoch uint64,
	endEpoch uint64,
) (*PerformanceSummary, error) {
	if endEpoch < startEpoch {
		return nil, fmt.Errorf("end epoch %d is before start epoch %d", endEpoch, startEpoch)
	}
	if cfg.SlotsPerEpoch == 0 {
		return nil, fmt.Errorf("invalid beacon config: SlotsPerEpoch is 0")
	}

	summary := &PerformanceSummary{
		ValidatorIndex:  validatorIndex,
		StartEpoch:      startEpoch,
		EndEpoch:        endEpoch,
		TotalEpochs:     endEpoch - startEpoch + 1,
		MissedEpochList: []uint64{},
		TimelyEpochList: []uint64{},
	}

	for epoch := startEpoch; epoch <= endEpoch; epoch++ {
		state, err := checkEpoch(bc, cfg, validatorIndex, epoch)
		if err != nil {
			return nil, fmt.Errorf("error checking epoch %d: %w", epoch, err)
		}
		switch state {
		case epochResultTimely:
			summary.TimelyEpochs++
			summary.TimelyEpochList = append(summary.TimelyEpochList, epoch)
		case epochResultMissed:
			summary.MissedEpochs++
			summary.MissedEpochList = append(summary.MissedEpochList, epoch)
		case epochResultInactive:
			summary.InactiveEpochs++
		}
	}

	activeEpochs := summary.TimelyEpochs + summary.MissedEpochs
	if activeEpochs > 0 {
		summary.PerformancePct = float64(summary.TimelyEpochs) / float64(activeEpochs) * 100.0
	}

	return summary, nil
}

// CheckEpochTargetVote returns true if the validator made a timely target
// vote for the given epoch. Returns (false, nil) for missed-target epochs;
// errors are reserved for I/O / parsing failures. If the validator was not
// in a committee for this epoch, the result is (true, nil) — there was no
// duty to perform, so it is not exit-eligible under RPIP-73.
//
// This is the lightweight single-epoch entry point intended for use by a
// challenge defender, who only needs to find one timely epoch within the
// challenged range.
func CheckEpochTargetVote(
	bc PerformanceBeaconClient,
	cfg beacon.Eth2Config,
	validatorIndex uint64,
	epoch uint64,
) (bool, error) {
	state, err := checkEpoch(bc, cfg, validatorIndex, epoch)
	if err != nil {
		return false, err
	}
	return state != epochResultMissed, nil
}

// epochResult enumerates the per-epoch verdicts.
type epochResult int

const (
	epochResultMissed epochResult = iota
	epochResultTimely
	epochResultInactive
)

// attestationDuty describes the unique (slot, committee, position) tuple
// assigned to a validator for an epoch.
type attestationDuty struct {
	slot                uint64
	committeeIndex      uint64
	position            int
	committeeSizesAtDay map[uint64]int // committee_index -> validator count, for all committees at duty.slot
}

// checkEpoch performs the per-epoch evaluation. It returns epochResultTimely
// if a matching, timely target vote was found for the validator;
// epochResultMissed if the validator had a duty but no matching attestation
// landed; epochResultInactive if the validator had no committee assignment
// for the epoch.
func checkEpoch(
	bc PerformanceBeaconClient,
	cfg beacon.Eth2Config,
	validatorIndex uint64,
	epoch uint64,
) (epochResult, error) {
	// Step 1: resolve the canonical target block root for the epoch. This is
	// the canonical block root at slot epoch*SlotsPerEpoch, walking back if
	// the boundary slot was skipped.
	targetRoot, err := resolveTargetRoot(bc, cfg, epoch)
	if err != nil {
		return epochResultMissed, err
	}

	// Step 2: find the validator's attestation duty for the epoch by walking
	// the committees response and matching against validatorIndex.
	duty, found, err := findAttestationDuty(bc, epoch, validatorIndex)
	if err != nil {
		return epochResultMissed, err
	}
	if !found {
		indexStr := strconv.FormatUint(validatorIndex, 10)
		status, err := bc.GetValidatorStatusByIndex(indexStr, nil)
		if err != nil {
			return epochResultMissed, fmt.Errorf("error getting validator status for index %d: %w", validatorIndex, err)
		}
		if !validatorHadAttestationDuty(status, epoch) {
			return epochResultInactive, nil
		}
		return epochResultMissed, fmt.Errorf(
			"validator index %d was active in epoch %d but was not found in attestation committees; ensure your beacon node provides historical committee data (archival node required)",
			validatorIndex, epoch,
		)
	}

	// Step 3: scan the inclusion window for a matching attestation. The
	// inclusion window for the target flag is up to SLOTS_PER_EPOCH slots
	// after the duty slot. Return as soon as we find a match.
	inclusionEndExclusive := duty.slot + 1 + cfg.SlotsPerEpoch
	for slot := duty.slot + 1; slot < inclusionEndExclusive; slot++ {
		block, exists, err := bc.GetBeaconBlock(strconv.FormatUint(slot, 10))
		if err != nil {
			return epochResultMissed, fmt.Errorf("error getting block at slot %d: %w", slot, err)
		}
		if !exists {
			continue
		}
		if matchesDuty(block.Attestations, duty, epoch, targetRoot) {
			return epochResultTimely, nil
		}
	}

	return epochResultMissed, nil
}

// resolveTargetRoot returns the canonical block root for epoch E, i.e.
// get_block_root(state, E) = canonical block root at slot E*SlotsPerEpoch
// (walking back if the boundary slot was skipped).
func resolveTargetRoot(bc PerformanceBeaconClient, cfg beacon.Eth2Config, epoch uint64) (common.Hash, error) {
	boundarySlot := epoch * cfg.SlotsPerEpoch
	for attempt := uint64(0); attempt < maxTargetRootWalkback; attempt++ {
		if attempt > boundarySlot {
			break
		}
		slot := boundarySlot - attempt
		header, exists, err := bc.GetBeaconBlockHeader(strconv.FormatUint(slot, 10))
		if err != nil {
			return common.Hash{}, fmt.Errorf("error getting beacon block header at slot %d: %w", slot, err)
		}
		if exists {
			return header.Root, nil
		}
	}
	return common.Hash{}, fmt.Errorf("could not find a non-skipped slot within %d slots of epoch %d boundary to resolve target root", maxTargetRootWalkback, epoch)
}

// findAttestationDuty walks the committees response for the epoch and
// returns the (slot, committee_index, position) assignment of the requested
// validator, plus the committee-size map for the duty slot (needed to compute
// the aggregation-bits offset for post-Electra attestations).
func findAttestationDuty(bc PerformanceBeaconClient, epoch uint64, validatorIndex uint64) (attestationDuty, bool, error) {
	committees, err := bc.GetCommitteesForEpoch(&epoch)
	if err != nil {
		return attestationDuty{}, false, fmt.Errorf("error getting committees for epoch %d: %w", epoch, err)
	}
	defer committees.Release()

	indexStr := strconv.FormatUint(validatorIndex, 10)

	// First pass: locate the validator's duty.
	var duty attestationDuty
	found := false
	for i := 0; i < committees.Count(); i++ {
		validators := committees.Validators(i)
		for pos, vIdx := range validators {
			if vIdx == indexStr {
				duty.slot = committees.Slot(i)
				duty.committeeIndex = committees.Index(i)
				duty.position = pos
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return attestationDuty{}, false, nil
	}

	// Second pass: collect committee sizes for all committees at duty.slot
	// so that ValidatorAttested can compute the correct aggregation-bits
	// offset under post-Electra attestation aggregation.
	duty.committeeSizesAtDay = map[uint64]int{}
	for i := 0; i < committees.Count(); i++ {
		if committees.Slot(i) != duty.slot {
			continue
		}
		duty.committeeSizesAtDay[committees.Index(i)] = committees.ValidatorCount(i)
	}

	return duty, true, nil
}

// matchesDuty returns true if any attestation in atts is a timely-target
// match for the given duty. The caller has already constrained inclusion
// delay by the slot range it iterated over.
func matchesDuty(atts []beacon.AttestationInfo, duty attestationDuty, epoch uint64, targetRoot common.Hash) bool {
	for _, att := range atts {
		if att.SlotIndex != duty.slot {
			continue
		}
		if att.TargetEpoch != epoch {
			continue
		}
		if att.TargetRoot != targetRoot {
			continue
		}
		// The attestation must cover the duty committee. Pre-Electra
		// attestations cover exactly one committee; post-Electra ones may
		// cover several. Either way, CommitteeIndices() returns the set.
		committeeIdxInt := int(duty.committeeIndex)
		hasCommittee := false
		for _, ci := range att.CommitteeIndices() {
			if ci == committeeIdxInt {
				hasCommittee = true
				break
			}
		}
		if !hasCommittee {
			continue
		}
		if att.ValidatorAttested(committeeIdxInt, duty.position, duty.committeeSizesAtDay) {
			return true
		}
	}
	return false
}

// validatorHadAttestationDuty reports whether the validator was required to
// perform an attestation duty in the given epoch based on its activation and
// exit epochs.
func validatorHadAttestationDuty(status beacon.ValidatorStatus, epoch uint64) bool {
	if !status.Exists {
		return false
	}
	if status.ActivationEpoch == farFutureEpoch || status.ActivationEpoch > epoch {
		return false
	}
	if status.ExitEpoch != farFutureEpoch && status.ExitEpoch <= epoch {
		return false
	}
	return true
}
