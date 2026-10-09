package recovery

import (
	"errors"
	"fmt"

	"github.com/rocket-pool/smartnode/shared/services/wallet"
)

// ErrUnresolvedKeysRefusingCommit is returned when default mode encounters unresolved or invalid keys.
var ErrUnresolvedKeysRefusingCommit = errors.New("refusing to commit recovery: plan contains unresolved or invalid validator keys")

// CommitPlan applies a pre-calculated RecoveryPlan using the provided KeyWriter.
// By default, if any in-scope key is unresolved or invalid, zero writes occur and an error is returned.
// When allowPartial is true, resolved keys are committed while unresolved ones are skipped and recorded.
func CommitPlan(plan *RecoveryPlan, writer KeyWriter, allowPartial bool) error {
	if err := validateCommitPreconditions(plan, writer, allowPartial); err != nil {
		return err
	}

	for i := range plan.Entries {
		if err := commitEntry(&plan.Entries[i], writer); err != nil {
			return err
		}
	}

	return nil
}

// validateCommitPreconditions verifies parameters and enforces the fail-fast safety invariant.
func validateCommitPreconditions(plan *RecoveryPlan, writer KeyWriter, allowPartial bool) error {
	if plan == nil {
		return errors.New("nil recovery plan provided")
	}
	if writer == nil {
		return errors.New("nil key writer provided")
	}

	if !allowPartial && plan.HasUnresolvedOrInvalid() {
		return fmt.Errorf("%w: %d unresolved, %d invalid (run with --allow-partial-recover to write only resolved keys)",
			ErrUnresolvedKeysRefusingCommit, plan.TotalUnresolved, plan.TotalInvalid)
	}

	return nil
}

// commitEntry routes key persistence or status updates for a single plan entry.
func commitEntry(entry *ValidatorPlanEntry, writer KeyWriter) error {
	switch entry.DiscoveryOutcome {
	case OutcomeExcludedInactive:
		entry.CommitOutcome = CommitSkippedExcluded
		return nil

	case OutcomeInstalledValid:
		entry.CommitOutcome = CommitSkippedAlreadyInstalled
		return nil

	case OutcomeUnresolved, OutcomeInvalidMaterial:
		entry.CommitOutcome = CommitSkippedPartial
		return nil

	case OutcomeMnemonicValid:
		return commitMnemonicEntry(entry, writer)

	case OutcomeCustomValid:
		return commitCustomEntry(entry, writer)

	default:
		entry.CommitOutcome = CommitNotAttempted
		return nil
	}
}

// commitMnemonicEntry validates and persists a mnemonic-derived validator key.
func commitMnemonicEntry(entry *ValidatorPlanEntry, writer KeyWriter) error {
	if entry.WalletIndex == nil || entry.PrivateKey == nil {
		entry.CommitOutcome = CommitFailed
		return fmt.Errorf("incomplete key material for mnemonic-derived pubkey %s", entry.Pubkey.Hex())
	}

	vk := wallet.ValidatorKey{
		PublicKey:      entry.Pubkey,
		PrivateKey:     entry.PrivateKey,
		DerivationPath: entry.DerivationPath,
		WalletIndex:    *entry.WalletIndex,
	}

	if err := writer.SaveValidatorKey(vk); err != nil {
		entry.CommitOutcome = CommitFailed
		return fmt.Errorf("error writing mnemonic validator key %s: %w", entry.Pubkey.Hex(), err)
	}

	entry.CommitOutcome = CommitWritten
	return nil
}

// commitCustomEntry validates and persists an imported custom validator keystore.
func commitCustomEntry(entry *ValidatorPlanEntry, writer KeyWriter) error {
	if entry.PrivateKey == nil {
		entry.CommitOutcome = CommitFailed
		return fmt.Errorf("incomplete key material for custom pubkey %s", entry.Pubkey.Hex())
	}

	if err := writer.StoreValidatorKey(entry.PrivateKey, entry.DerivationPath); err != nil {
		entry.CommitOutcome = CommitFailed
		return fmt.Errorf("error storing custom validator key %s: %w", entry.Pubkey.Hex(), err)
	}

	entry.CommitOutcome = CommitWritten
	return nil
}
