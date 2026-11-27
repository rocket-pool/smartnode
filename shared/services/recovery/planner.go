package recovery

import (
	"bytes"
	"fmt"

	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
)

// Default limits matching Smartnode recovery constants.
const (
	DefaultBucketSize  uint = 20
	DefaultBucketLimit uint = 2000
)

// PlannerOptions controls configuration for key discovery.
type PlannerOptions struct {
	BucketSize  uint
	BucketLimit uint
}

// DefaultPlannerOptions returns sensible default planner settings.
func DefaultPlannerOptions() PlannerOptions {
	return PlannerOptions{
		BucketSize:  DefaultBucketSize,
		BucketLimit: DefaultBucketLimit,
	}
}

// PlanValidatorRecovery builds a complete, deterministic RecoveryPlan without mutating any files.
func PlanValidatorRecovery(
	pubkeys []types.ValidatorPubkey,
	statuses map[types.ValidatorPubkey]beacon.ValidatorStatus,
	deriver KeyDeriver,
	installedChecker InstalledKeyChecker,
	customProvider CustomKeyProvider,
	opts PlannerOptions,
) (*RecoveryPlan, error) {
	opts = normalizePlannerOptions(opts)

	validPubkeys := filterAndDeduplicatePubkeys(pubkeys)
	entryMap, unresolvedMap := initializePlanEntries(validPubkeys, statuses)

	checkInstalledKeys(installedChecker, unresolvedMap)

	if err := inspectCustomKeys(customProvider, unresolvedMap); err != nil {
		return nil, err
	}

	if err := searchMnemonicKeys(deriver, unresolvedMap, opts); err != nil {
		return nil, err
	}

	return buildRecoveryPlan(validPubkeys, entryMap, unresolvedMap, opts.BucketLimit), nil
}

// normalizePlannerOptions ensures option fields have safe non-zero defaults.
func normalizePlannerOptions(opts PlannerOptions) PlannerOptions {
	if opts.BucketSize == 0 {
		opts.BucketSize = DefaultBucketSize
	}
	if opts.BucketLimit == 0 {
		opts.BucketLimit = DefaultBucketLimit
	}
	return opts
}

// filterAndDeduplicatePubkeys strips out empty/zero pubkeys and duplicate entries, preserving order.
func filterAndDeduplicatePubkeys(pubkeys []types.ValidatorPubkey) []types.ValidatorPubkey {
	zeroPubkey := types.ValidatorPubkey{}
	seen := make(map[types.ValidatorPubkey]bool, len(pubkeys))
	valid := make([]types.ValidatorPubkey, 0, len(pubkeys))

	for _, pk := range pubkeys {
		if bytes.Equal(pk[:], zeroPubkey[:]) || seen[pk] {
			continue
		}
		seen[pk] = true
		valid = append(valid, pk)
	}
	return valid
}

// initializePlanEntries creates entries for all pubkeys and identifies inactive beacon validators.
// Fixes Issue #407: inactive validators are categorized as OutcomeExcludedInactive.
func initializePlanEntries(
	validPubkeys []types.ValidatorPubkey,
	statuses map[types.ValidatorPubkey]beacon.ValidatorStatus,
) (map[types.ValidatorPubkey]*ValidatorPlanEntry, map[types.ValidatorPubkey]*ValidatorPlanEntry) {
	entryMap := make(map[types.ValidatorPubkey]*ValidatorPlanEntry, len(validPubkeys))
	unresolvedMap := make(map[types.ValidatorPubkey]*ValidatorPlanEntry)

	for _, pk := range validPubkeys {
		entry := &ValidatorPlanEntry{
			Pubkey:        pk,
			Source:        SourceNone,
			CommitOutcome: CommitNotAttempted,
		}
		entryMap[pk] = entry

		if statuses != nil {
			if status, hasStatus := statuses[pk]; hasStatus && !isActiveValidatorState(status.Status) {
				entry.DiscoveryOutcome = OutcomeExcludedInactive
				entry.CommitOutcome = CommitSkippedExcluded
				entry.Details = fmt.Sprintf("beacon validator state: %s", status.Status)
				continue
			}
		}

		entry.DiscoveryOutcome = OutcomeUnresolved
		unresolvedMap[pk] = entry
	}

	return entryMap, unresolvedMap
}

// checkInstalledKeys identifies validator keys that are already installed in validator clients.
func checkInstalledKeys(
	installedChecker InstalledKeyChecker,
	unresolvedMap map[types.ValidatorPubkey]*ValidatorPlanEntry,
) {
	if installedChecker == nil || len(unresolvedMap) == 0 {
		return
	}

	for pk, entry := range unresolvedMap {
		installed, err := installedChecker.IsKeyInstalled(pk)
		if err == nil && installed {
			entry.Source = SourceInstalled
			entry.DiscoveryOutcome = OutcomeInstalledValid
			entry.CommitOutcome = CommitSkippedAlreadyInstalled
			entry.Details = "key material already present in validator client keystores"
			delete(unresolvedMap, pk)
		}
	}
}

// inspectCustomKeys processes imported EIP-2335 custom keystores without writing anything to disk.
func inspectCustomKeys(
	customProvider CustomKeyProvider,
	unresolvedMap map[types.ValidatorPubkey]*ValidatorPlanEntry,
) error {
	if customProvider == nil || len(unresolvedMap) == 0 {
		return nil
	}

	customKeys, err := customProvider.GetCustomKeys()
	if err != nil {
		return fmt.Errorf("error loading custom keys: %w", err)
	}

	for _, ck := range customKeys {
		entry, exists := unresolvedMap[ck.Pubkey]
		if !exists {
			continue
		}

		entry.Source = SourceCustomKeystore
		if ck.Error != nil {
			entry.DiscoveryOutcome = OutcomeInvalidMaterial
			entry.CommitOutcome = CommitNotAttempted
			entry.Details = ck.Error.Error()
		} else {
			entry.DiscoveryOutcome = OutcomeCustomValid
			entry.CommitOutcome = CommitNotAttempted
			entry.PrivateKey = ck.PrivateKey
			entry.DerivationPath = ck.DerivationPath
			entry.Details = fmt.Sprintf("loaded from custom keystore: %s", ck.FileName)
		}
		delete(unresolvedMap, ck.Pubkey)
	}

	return nil
}

// searchMnemonicKeys derives keys from the wallet mnemonic seed within the configured bucket limits.
func searchMnemonicKeys(
	deriver KeyDeriver,
	unresolvedMap map[types.ValidatorPubkey]*ValidatorPlanEntry,
	opts PlannerOptions,
) error {
	if deriver == nil || len(unresolvedMap) == 0 {
		return nil
	}

	bucketStart := uint(0)
	for bucketStart < opts.BucketLimit && len(unresolvedMap) > 0 {
		bucketEnd := bucketStart + opts.BucketSize
		if bucketEnd > opts.BucketLimit {
			bucketEnd = opts.BucketLimit
		}

		keys, err := deriver.GetValidatorKeys(bucketStart, bucketEnd-bucketStart)
		if err != nil {
			return fmt.Errorf("error deriving validator keys at index %d: %w", bucketStart, err)
		}

		for _, vk := range keys {
			entry, exists := unresolvedMap[vk.PublicKey]
			if exists {
				wIndex := vk.WalletIndex
				entry.Source = SourceMnemonic
				entry.DiscoveryOutcome = OutcomeMnemonicValid
				entry.CommitOutcome = CommitNotAttempted
				entry.PrivateKey = vk.PrivateKey
				entry.DerivationPath = vk.DerivationPath
				entry.WalletIndex = &wIndex
				entry.Details = fmt.Sprintf("derived at index %d (%s)", vk.WalletIndex, vk.DerivationPath)
				delete(unresolvedMap, vk.PublicKey)
			}
		}

		bucketStart = bucketEnd
	}

	return nil
}

// buildRecoveryPlan sets final details for remaining unresolved keys and packages the plan.
func buildRecoveryPlan(
	validPubkeys []types.ValidatorPubkey,
	entryMap map[types.ValidatorPubkey]*ValidatorPlanEntry,
	unresolvedMap map[types.ValidatorPubkey]*ValidatorPlanEntry,
	bucketLimit uint,
) *RecoveryPlan {
	for _, entry := range unresolvedMap {
		entry.Details = fmt.Sprintf("not resolved within derivation limit (%d attempts)", bucketLimit)
	}

	plan := NewRecoveryPlan()
	for _, pk := range validPubkeys {
		plan.AddEntry(*entryMap[pk])
	}
	return plan
}

// isActiveValidatorState checks if a beacon validator status is in an active or pending state.
func isActiveValidatorState(state beacon.ValidatorState) bool {
	return state == beacon.ValidatorState_ActiveOngoing ||
		state == beacon.ValidatorState_ActiveExiting ||
		state == beacon.ValidatorState_PendingInitialized ||
		state == beacon.ValidatorState_PendingQueued
}
