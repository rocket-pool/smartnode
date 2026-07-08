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
package performance

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/rocket-pool/smartnode/bindings/rocketpool"
	"github.com/rocket-pool/smartnode/bindings/settings/protocol"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/bindings/utils/eth"

	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/state"
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

// defaultPerformanceThresholdPct is the RPIP-73 initial pDAO value
const defaultPerformanceThresholdPct = 94.0

func GetPerformanceThresholdPct(rp *rocketpool.RocketPool) (float64, error) {
	saturn2Deployed, err := state.IsSaturn2Deployed(rp, nil)
	if err != nil {
		return 0, fmt.Errorf("error checking if Saturn 2 is deployed: %w", err)
	}
	if !saturn2Deployed {
		return defaultPerformanceThresholdPct, nil
	}
	thresholdWei, err := protocol.GetPerformanceThreshold(rp, nil)
	if err != nil {
		return 0, fmt.Errorf("error getting performance threshold: %w", err)
	}
	return eth.WeiToEth(thresholdWei) * 100.0, nil
}

// Defaults used before Saturn 2 deploys.
const DefaultPerformancePeriodEpochs uint64 = 44032
const defaultProofBuffer = 24 * time.Hour

// ChallengeParams are the pDAO settings governing performance challenges.
type ChallengeParams struct {
	ExitsEnabled bool
	PeriodEpochs uint64
	ProofBuffer  time.Duration
}

// GetChallengeParams fetches the pDAO performance-challenge settings, using
// the pre-Saturn-2 defaults when Saturn 2 is not deployed yet.
func GetChallengeParams(rp *rocketpool.RocketPool) (ChallengeParams, error) {
	saturn2Deployed, err := state.IsSaturn2Deployed(rp, nil)
	if err != nil {
		return ChallengeParams{}, fmt.Errorf("error checking if Saturn 2 is deployed: %w", err)
	}
	if !saturn2Deployed {
		return ChallengeParams{
			ExitsEnabled: true,
			PeriodEpochs: DefaultPerformancePeriodEpochs,
			ProofBuffer:  defaultProofBuffer,
		}, nil
	}
	exitsEnabled, err := protocol.GetPerformanceExitsEnabled(rp, nil)
	if err != nil {
		return ChallengeParams{}, err
	}
	periodEpochs, err := protocol.GetPerformancePeriod(rp, nil)
	if err != nil {
		return ChallengeParams{}, err
	}
	proofBuffer, err := protocol.GetPerformanceProofBuffer(rp, nil)
	if err != nil {
		return ChallengeParams{}, err
	}
	return ChallengeParams{
		ExitsEnabled: exitsEnabled,
		PeriodEpochs: periodEpochs,
		ProofBuffer:  proofBuffer,
	}, nil
}

// challengeBeaconClient is the beacon client surface needed to evaluate the
// challengeability of an epoch range.
type challengeBeaconClient interface {
	GetEth2Config() (beacon.Eth2Config, error)
	GetBeaconHead() (beacon.BeaconHead, error)
}

// IsRangeChallengeable fetches the pDAO challenge settings and the beacon
// head, then reports whether a performance check over the inclusive range
// [startEpoch, endEpoch] could back an on-chain challenge. See
// IsChallengeable for the rules.
func IsRangeChallengeable(rp *rocketpool.RocketPool, bc challengeBeaconClient, startEpoch, endEpoch uint64) (bool, error) {
	params, err := GetChallengeParams(rp)
	if err != nil {
		return false, err
	}
	cfg, err := bc.GetEth2Config()
	if err != nil {
		return false, fmt.Errorf("error getting beacon config: %w", err)
	}
	head, err := bc.GetBeaconHead()
	if err != nil {
		return false, fmt.Errorf("error getting beacon head: %w", err)
	}
	return IsChallengeable(params, cfg, head.Epoch, startEpoch, endEpoch), nil
}

// ExceedsChallengeThreshold reports whether the validator missed enough
// target votes for a challenge to succeed: the missed share of the checked
// period must be higher than the allowed slack (100% - performance_threshold).
func ExceedsChallengeThreshold(resp *api.VerifyPerformanceResponse) bool {
	if resp.TotalEpochs == 0 {
		return false
	}
	missedPct := float64(resp.MissedEpochs) / float64(resp.TotalEpochs) * 100.0
	return missedPct > 100.0-resp.PerformanceThresholdPct
}

// IsChallengeable reports whether a performance check over the inclusive
// range [startEpoch, endEpoch] could back an on-chain challenge: performance
// exits must be enabled, the range must cover exactly one performance period,
// and it must be recent enough that the proof buffer has not elapsed
// (startEpoch > currentEpoch - period - proofBuffer, with the buffer
// converted to epochs).
func IsChallengeable(params ChallengeParams, cfg beacon.Eth2Config, currentEpoch, startEpoch, endEpoch uint64) bool {
	if !params.ExitsEnabled {
		return false
	}
	if endEpoch != startEpoch+params.PeriodEpochs-1 {
		return false
	}
	if cfg.SecondsPerEpoch == 0 {
		return false
	}
	proofBufferEpochs := uint64(params.ProofBuffer.Seconds()) / cfg.SecondsPerEpoch
	window := params.PeriodEpochs + proofBufferEpochs
	if currentEpoch <= window {
		// The whole chain history is still within the challenge window.
		return true
	}
	return startEpoch > currentEpoch-window
}

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

// pubkeyBeaconClient is the beacon client surface needed to resolve validator
// pubkeys to beacon-chain indices in addition to the engine requirements.
type pubkeyBeaconClient interface {
	PerformanceBeaconClient
	GetValidatorIndex(pubkey rptypes.ValidatorPubkey) (string, error)
	GetValidatorStatuses(pubkeys []rptypes.ValidatorPubkey, opts *beacon.ValidatorStatusOptions) (map[rptypes.ValidatorPubkey]beacon.ValidatorStatus, error)
}

// VerifyPerformance is the end-to-end RPIP-73 target-vote verification flow for
// a single validator. It resolves the validator's beacon-chain index from the
// supplied pubkey, runs CheckTargetPerformance, and packages the result
// alongside the pDAO performance_threshold for pass/fail reporting.
//
// To verify several validators in one run, prefer VerifyPerformanceBatch, which
// shares per-epoch beacon data across all of them.
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

	thresholdPct, err := GetPerformanceThresholdPct(rp)
	if err != nil {
		return nil, err
	}

	summary, err := CheckTargetPerformance(bc, cfg, validatorIndex, startEpoch, endEpoch)
	if err != nil {
		return nil, err
	}

	return summaryToResponse(pubkey, summary, thresholdPct), nil
}

// BatchValidatorResult is one validator's outcome from VerifyPerformanceBatch.
// Exactly one of Response or Err is set: Response when the check succeeded, Err
// when that single validator could not be verified (the rest of the batch is
// unaffected). The slice returned by VerifyPerformanceBatch is aligned
// positionally with the input pubkeys, so callers can map results back to their
// own identifiers (minipool address, megapool validator id, etc).
type BatchValidatorResult struct {
	Pubkey rptypes.ValidatorPubkey
	// Active reports whether the validator is currently active on the beacon
	// chain (activated and not yet exited) according to its live head status.
	// Callers that target "all" validators use this to skip validators that are
	// not actively attesting.
	Active   bool
	Response *api.VerifyPerformanceResponse
	Err      error
}

// isActiveValidatorState reports whether a beacon validator state counts as
// actively attesting (activated on the beacon chain and not yet exited).
func isActiveValidatorState(state beacon.ValidatorState) bool {
	switch state {
	case beacon.ValidatorState_ActiveOngoing,
		beacon.ValidatorState_ActiveExiting,
		beacon.ValidatorState_ActiveSlashed:
		return true
	default:
		return false
	}
}

// VerifyPerformanceBatch verifies the RPIP-73 target-vote performance of many
// validators over the same inclusive epoch range in a single pass. All
// per-epoch beacon data (target roots, committee assignments, inclusion-window
// blocks) is fetched once via a shared epochCache and reused for every
// validator, and the validators' indices and statuses are resolved in a single
// batched beacon call.
//
// The returned slice is positionally aligned with pubkeys. A fatal error
// (returned as the second value) only occurs for failures that prevent the
// whole batch from running, such as being unable to read the beacon config or
// resolve any validator statuses; per-validator problems are reported in each
// entry's Err field instead.
func VerifyPerformanceBatch(
	rp *rocketpool.RocketPool,
	bc pubkeyBeaconClient,
	pubkeys []rptypes.ValidatorPubkey,
	startEpoch uint64,
	endEpoch uint64,
) ([]BatchValidatorResult, error) {
	if endEpoch < startEpoch {
		return nil, fmt.Errorf("end epoch %d is before start epoch %d", endEpoch, startEpoch)
	}

	cfg, err := bc.GetEth2Config()
	if err != nil {
		return nil, fmt.Errorf("error getting beacon config: %w", err)
	}
	if cfg.SlotsPerEpoch == 0 {
		return nil, fmt.Errorf("invalid beacon config: SlotsPerEpoch is 0")
	}

	// Fetch the pDAO performance threshold once for the whole batch.
	thresholdPct, err := GetPerformanceThresholdPct(rp)
	if err != nil {
		return nil, err
	}

	// Resolve every non-zero pubkey to its index and status in a single call.
	uniquePubkeys := make([]rptypes.ValidatorPubkey, 0, len(pubkeys))
	seen := map[rptypes.ValidatorPubkey]struct{}{}
	for _, pk := range pubkeys {
		if pk == (rptypes.ValidatorPubkey{}) {
			continue
		}
		if _, ok := seen[pk]; ok {
			continue
		}
		seen[pk] = struct{}{}
		uniquePubkeys = append(uniquePubkeys, pk)
	}

	statusByPubkey := map[rptypes.ValidatorPubkey]beacon.ValidatorStatus{}
	if len(uniquePubkeys) > 0 {
		statusByPubkey, err = bc.GetValidatorStatuses(uniquePubkeys, nil)
		if err != nil {
			return nil, fmt.Errorf("error resolving validator statuses: %w", err)
		}
	}

	// Build the set of indices to track and a pre-populated status map keyed by
	// index string so the cache never has to re-query a status.
	indexSet := map[string]struct{}{}
	statusByIndex := map[string]beacon.ValidatorStatus{}
	for _, status := range statusByPubkey {
		if !status.Exists || status.Index == "" {
			continue
		}
		indexSet[status.Index] = struct{}{}
		statusByIndex[status.Index] = status
	}

	cache := newEpochCache(bc, cfg, indexSet, statusByIndex)

	results := make([]BatchValidatorResult, len(pubkeys))
	for i, pk := range pubkeys {
		results[i].Pubkey = pk
		if pk == (rptypes.ValidatorPubkey{}) {
			results[i].Err = fmt.Errorf("validator has no pubkey on-chain yet (not deposited?)")
			continue
		}
		status, ok := statusByPubkey[pk]
		if !ok || !status.Exists || status.Index == "" {
			results[i].Err = fmt.Errorf("validator %s not found on the beacon chain yet", pk.Hex())
			continue
		}
		results[i].Active = isActiveValidatorState(status.Status)
		indexU64, err := strconv.ParseUint(status.Index, 10, 64)
		if err != nil {
			results[i].Err = fmt.Errorf("error parsing validator index %q: %w", status.Index, err)
			continue
		}
		summary, err := cache.computeSummary(status.Index, indexU64, startEpoch, endEpoch)
		if err != nil {
			results[i].Err = err
			continue
		}
		results[i].Response = summaryToResponse(pk, summary, thresholdPct)
	}

	return results, nil
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

	indexStr := strconv.FormatUint(validatorIndex, 10)
	cache := newEpochCache(bc, cfg, map[string]struct{}{indexStr: {}}, nil)
	return cache.computeSummary(indexStr, validatorIndex, startEpoch, endEpoch)
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
	indexStr := strconv.FormatUint(validatorIndex, 10)
	cache := newEpochCache(bc, cfg, map[string]struct{}{indexStr: {}}, nil)
	state, err := cache.evaluateEpoch(indexStr, validatorIndex, epoch)
	if err != nil {
		return false, err
	}
	return state != epochResultMissed, nil
}

// FindFirstTimelyTargetVote scans the supplied validator indices over the
// supplied epochs and returns the first validator index (and the epoch) that
// made a valid target vote
func FindFirstTimelyTargetVote(
	bc PerformanceBeaconClient,
	cfg beacon.Eth2Config,
	validatorIndices []uint64,
	epochs []uint64,
) (validatorIndex uint64, epoch uint64, found bool, err error) {
	if cfg.SlotsPerEpoch == 0 {
		return 0, 0, false, fmt.Errorf("invalid beacon config: SlotsPerEpoch is 0")
	}

	// Build the tracked index set (keyed by index string, as the cache keys on
	// those) while de-duplicating the input.
	indexSet := make(map[string]struct{}, len(validatorIndices))
	indexStrByIndex := make(map[uint64]string, len(validatorIndices))
	uniqueIndices := make([]uint64, 0, len(validatorIndices))
	for _, idx := range validatorIndices {
		if _, seen := indexStrByIndex[idx]; seen {
			continue
		}
		s := strconv.FormatUint(idx, 10)
		indexSet[s] = struct{}{}
		indexStrByIndex[idx] = s
		uniqueIndices = append(uniqueIndices, idx)
	}

	cache := newEpochCache(bc, cfg, indexSet, nil)

	// Inspect each validator over the challenged epochs, returning as soon as
	// one of them is found to have made a timely target vote.
	for _, idx := range uniqueIndices {
		for _, ep := range epochs {
			state, err := cache.evaluateEpoch(indexStrByIndex[idx], idx, ep)
			if err != nil {
				return 0, 0, false, fmt.Errorf("error checking validator index %d in epoch %d: %w", idx, ep, err)
			}
			if state == epochResultTimely {
				return idx, ep, true, nil
			}
		}
	}

	return 0, 0, false, nil
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

// cachedBlock memoizes a single GetBeaconBlock lookup, including the "missing"
// (skipped slot) case.
type cachedBlock struct {
	block  beacon.BeaconBlock
	exists bool
}

// rootResult memoizes a target-root resolution, including its error.
type rootResult struct {
	root common.Hash
	err  error
}

// epochCache fetches and memoizes the per-epoch beacon data needed for
// target-vote verification so it can be shared across many validators in a
// single run. It is not safe for concurrent use.
type epochCache struct {
	bc       PerformanceBeaconClient
	cfg      beacon.Eth2Config
	indexSet map[string]struct{}

	targetRoots map[uint64]rootResult                 // epoch -> target root
	epochDuties map[uint64]map[string]attestationDuty // epoch -> validator index string -> duty
	blocks      map[uint64]cachedBlock                // slot -> block
	statuses    map[string]beacon.ValidatorStatus     // validator index string -> status
}

// newEpochCache creates a cache that tracks the supplied validator index set.
// statuses may be nil; when provided it pre-populates validator statuses (keyed
// by index string) so the cache never has to query them again for inactive
// detection.
func newEpochCache(
	bc PerformanceBeaconClient,
	cfg beacon.Eth2Config,
	indexSet map[string]struct{},
	statuses map[string]beacon.ValidatorStatus,
) *epochCache {
	if statuses == nil {
		statuses = map[string]beacon.ValidatorStatus{}
	}
	return &epochCache{
		bc:          bc,
		cfg:         cfg,
		indexSet:    indexSet,
		targetRoots: map[uint64]rootResult{},
		epochDuties: map[uint64]map[string]attestationDuty{},
		blocks:      map[uint64]cachedBlock{},
		statuses:    statuses,
	}
}

// computeSummary evaluates one validator over the inclusive epoch range using
// the shared cache.
func (c *epochCache) computeSummary(indexStr string, indexU64 uint64, startEpoch, endEpoch uint64) (*PerformanceSummary, error) {
	summary := &PerformanceSummary{
		ValidatorIndex:  indexU64,
		StartEpoch:      startEpoch,
		EndEpoch:        endEpoch,
		TotalEpochs:     endEpoch - startEpoch + 1,
		MissedEpochList: []uint64{},
		TimelyEpochList: []uint64{},
	}

	for epoch := startEpoch; epoch <= endEpoch; epoch++ {
		state, err := c.evaluateEpoch(indexStr, indexU64, epoch)
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

// evaluateEpoch performs the per-epoch evaluation for a single validator. It
// returns epochResultTimely if a matching, timely target vote was found;
// epochResultMissed if the validator had a duty but no matching attestation
// landed; epochResultInactive if the validator had no committee assignment for
// the epoch (and was not required to attest).
func (c *epochCache) evaluateEpoch(indexStr string, indexU64 uint64, epoch uint64) (epochResult, error) {
	// Find the validator's attestation duty for the epoch from the (cached)
	// committees.
	duty, found, err := c.dutyFor(epoch, indexStr)
	if err != nil {
		return epochResultMissed, err
	}
	if !found {
		status, err := c.status(indexStr)
		if err != nil {
			return epochResultMissed, fmt.Errorf("error getting validator status for index %s: %w", indexStr, err)
		}
		if !validatorHadAttestationDuty(status, epoch) {
			return epochResultInactive, nil
		}
		return epochResultMissed, fmt.Errorf(
			"validator index %s was active in epoch %d but was not found in attestation committees; ensure your beacon node provides historical committee data (archival node required)",
			indexStr, epoch,
		)
	}

	// Resolve the canonical target block root for the epoch (cached, shared).
	targetRoot, err := c.targetRoot(epoch)
	if err != nil {
		return epochResultMissed, err
	}

	// Scan the inclusion window for a matching attestation. The inclusion
	// window for the target flag is up to SLOTS_PER_EPOCH slots after the duty
	// slot. Return as soon as we find a match.
	inclusionEndExclusive := duty.slot + 1 + c.cfg.SlotsPerEpoch
	for slot := duty.slot + 1; slot < inclusionEndExclusive; slot++ {
		block, exists, err := c.block(slot)
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

// targetRoot returns the canonical block root for epoch E, memoized per epoch.
func (c *epochCache) targetRoot(epoch uint64) (common.Hash, error) {
	if r, ok := c.targetRoots[epoch]; ok {
		return r.root, r.err
	}
	root, err := resolveTargetRoot(c.bc, c.cfg, epoch)
	c.targetRoots[epoch] = rootResult{root: root, err: err}
	return root, err
}

// block returns the beacon block at the given slot, memoized per slot. The
// returned bool reports whether a block exists at that slot (false for a
// skipped slot).
func (c *epochCache) block(slot uint64) (beacon.BeaconBlock, bool, error) {
	if b, ok := c.blocks[slot]; ok {
		return b.block, b.exists, nil
	}
	block, exists, err := c.bc.GetBeaconBlock(strconv.FormatUint(slot, 10))
	if err != nil {
		return beacon.BeaconBlock{}, false, err
	}
	c.blocks[slot] = cachedBlock{block: block, exists: exists}
	return block, exists, nil
}

// status returns the validator status for the given index string, memoized and
// using any pre-populated statuses first.
func (c *epochCache) status(indexStr string) (beacon.ValidatorStatus, error) {
	if status, ok := c.statuses[indexStr]; ok {
		return status, nil
	}
	status, err := c.bc.GetValidatorStatusByIndex(indexStr, nil)
	if err != nil {
		return beacon.ValidatorStatus{}, err
	}
	c.statuses[indexStr] = status
	return status, nil
}

// dutyFor returns the requested validator's attestation duty for the epoch,
// building (and caching) the duty map for all tracked indices on first access.
func (c *epochCache) dutyFor(epoch uint64, indexStr string) (attestationDuty, bool, error) {
	if err := c.ensureEpochDuties(epoch); err != nil {
		return attestationDuty{}, false, err
	}
	duty, found := c.epochDuties[epoch][indexStr]
	return duty, found, nil
}

// ensureEpochDuties fetches the committees for the epoch once and, in a single
// pass, records the attestation duties of every tracked validator index along
// with the committee sizes needed to compute aggregation-bits offsets.
func (c *epochCache) ensureEpochDuties(epoch uint64) error {
	if _, ok := c.epochDuties[epoch]; ok {
		return nil
	}

	committees, err := c.bc.GetCommitteesForEpoch(&epoch)
	if err != nil {
		return fmt.Errorf("error getting committees for epoch %d: %w", epoch, err)
	}
	defer committees.Release()

	duties := map[string]attestationDuty{}
	// slot -> committee_index -> validator count, for every committee in the
	// epoch. Needed because post-Electra aggregation_bits pack multiple
	// committees per slot, so the offset of a validator depends on the sizes of
	// the lower-indexed committees at its slot.
	slotSizes := map[uint64]map[uint64]int{}

	for i := 0; i < committees.Count(); i++ {
		slot := committees.Slot(i)
		committeeIndex := committees.Index(i)
		validators := committees.Validators(i)

		if slotSizes[slot] == nil {
			slotSizes[slot] = map[uint64]int{}
		}
		slotSizes[slot][committeeIndex] = len(validators)

		// Only bother matching when there are still tracked validators left to
		// find that could be in this committee.
		for pos, vIdx := range validators {
			if _, tracked := c.indexSet[vIdx]; !tracked {
				continue
			}
			if _, already := duties[vIdx]; already {
				continue
			}
			duties[vIdx] = attestationDuty{
				slot:           slot,
				committeeIndex: committeeIndex,
				position:       pos,
			}
		}
	}

	// Attach the committee-size map for each duty's slot.
	for vIdx, duty := range duties {
		duty.committeeSizesAtDay = slotSizes[duty.slot]
		duties[vIdx] = duty
	}

	c.epochDuties[epoch] = duties
	return nil
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

// bitsPerParticipationWord matches the Solidity uint256 word width
const bitsPerParticipationWord = 256

// EncodeParticipationBitset encodes the missed epochs of the inclusive range
// [startEpoch, endEpoch] as the uint256[] bitset expected by the
// challengeMegapool participation calldata: the words form a single bit
// stream starting at startEpoch, LSB-first within each word, with a 1 bit
// marking a not-timely target attestation.
func EncodeParticipationBitset(startEpoch, endEpoch uint64, missedEpochs []uint64) []*big.Int {
	if endEpoch < startEpoch {
		return []*big.Int{}
	}
	totalEpochs := endEpoch - startEpoch + 1
	wordCount := (totalEpochs + bitsPerParticipationWord - 1) / bitsPerParticipationWord
	words := make([]*big.Int, wordCount)
	for i := range words {
		words[i] = new(big.Int)
	}
	for _, epoch := range missedEpochs {
		if epoch < startEpoch || epoch > endEpoch {
			continue
		}
		offset := epoch - startEpoch
		word := words[offset/bitsPerParticipationWord]
		word.SetBit(word, int(offset%bitsPerParticipationWord), 1)
	}
	return words
}

// summaryToResponse packages a PerformanceSummary into the API response,
// attaching the pDAO performance_threshold and pass/fail verdict.
func summaryToResponse(pubkey rptypes.ValidatorPubkey, summary *PerformanceSummary, thresholdPct float64) *api.VerifyPerformanceResponse {
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
		Participation:           EncodeParticipationBitset(summary.StartEpoch, summary.EndEpoch, summary.MissedEpochList),
	}
}
