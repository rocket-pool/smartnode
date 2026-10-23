package recovery

import (
	"encoding/json"
	"fmt"

	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
)

// KeyDiscoverySource identifies how key material for a validator was obtained.
type KeyDiscoverySource string

const (
	SourceNone           KeyDiscoverySource = "none"
	SourceMnemonic       KeyDiscoverySource = "mnemonic"
	SourceCustomKeystore KeyDiscoverySource = "custom_keystore"
	SourceInstalled      KeyDiscoverySource = "installed"
)

// DiscoveryOutcome represents the deterministic diagnosis of a validator pubkey during planning.
type DiscoveryOutcome string

const (
	OutcomeMnemonicValid    DiscoveryOutcome = "mnemonic_valid"
	OutcomeCustomValid      DiscoveryOutcome = "custom_valid"
	OutcomeInstalledValid   DiscoveryOutcome = "installed_valid"
	OutcomeUnresolved       DiscoveryOutcome = "unresolved"
	OutcomeInvalidMaterial  DiscoveryOutcome = "invalid_material"
	OutcomeExcludedInactive DiscoveryOutcome = "excluded_inactive"
)

// CommitOutcome represents the result of attempting to write key material to disk.
type CommitOutcome string

const (
	CommitNotAttempted            CommitOutcome = "not_attempted"
	CommitWritten                 CommitOutcome = "written"
	CommitSkippedAlreadyInstalled CommitOutcome = "skipped_already_installed"
	CommitSkippedExcluded         CommitOutcome = "skipped_excluded"
	CommitSkippedPartial          CommitOutcome = "skipped_partial"
	CommitFailed                  CommitOutcome = "failed"
)

// ValidatorPlanEntry contains the discovery diagnosis and commit result for a single validator.
type ValidatorPlanEntry struct {
	Pubkey           types.ValidatorPubkey `json:"pubkey"`
	Source           KeyDiscoverySource    `json:"source"`
	DiscoveryOutcome DiscoveryOutcome      `json:"discoveryOutcome"`
	CommitOutcome    CommitOutcome         `json:"commitOutcome"`
	Details          string                `json:"details,omitempty"`
	DerivationPath   string                `json:"derivationPath,omitempty"`
	WalletIndex      *uint                 `json:"walletIndex,omitempty"`

	// In-memory key material kept only during planning/committing and never serialized.
	PrivateKey *eth2types.BLSPrivateKey `json:"-"`
}

// IsResolved returns true if the entry was successfully discovered and has valid key material.
func (e *ValidatorPlanEntry) IsResolved() bool {
	return e.DiscoveryOutcome == OutcomeMnemonicValid ||
		e.DiscoveryOutcome == OutcomeCustomValid ||
		e.DiscoveryOutcome == OutcomeInstalledValid
}

// IsExcluded returns true if the validator was excluded due to inactive beacon chain state.
func (e *ValidatorPlanEntry) IsExcluded() bool {
	return e.DiscoveryOutcome == OutcomeExcludedInactive
}

// RecoveryPlan encapsulates the entire recovery state across all in-scope validators.
type RecoveryPlan struct {
	Entries         []ValidatorPlanEntry `json:"entries"`
	TotalInScope    int                  `json:"totalInScope"`
	TotalResolved   int                  `json:"totalResolved"`
	TotalUnresolved int                  `json:"totalUnresolved"`
	TotalInvalid    int                  `json:"totalInvalid"`
	TotalExcluded   int                  `json:"totalExcluded"`
}

// NewRecoveryPlan constructs an initialized RecoveryPlan.
func NewRecoveryPlan() *RecoveryPlan {
	return &RecoveryPlan{
		Entries: []ValidatorPlanEntry{},
	}
}

// AddEntry appends an entry to the plan and updates summary counts.
func (p *RecoveryPlan) AddEntry(entry ValidatorPlanEntry) {
	p.Entries = append(p.Entries, entry)
	p.recalculateTotals()
}

func (p *RecoveryPlan) recalculateTotals() {
	p.TotalInScope = 0
	p.TotalResolved = 0
	p.TotalUnresolved = 0
	p.TotalInvalid = 0
	p.TotalExcluded = 0

	for _, e := range p.Entries {
		switch e.DiscoveryOutcome {
		case OutcomeExcludedInactive:
			p.TotalExcluded++
		case OutcomeMnemonicValid, OutcomeCustomValid, OutcomeInstalledValid:
			p.TotalInScope++
			p.TotalResolved++
		case OutcomeUnresolved:
			p.TotalInScope++
			p.TotalUnresolved++
		case OutcomeInvalidMaterial:
			p.TotalInScope++
			p.TotalInvalid++
		}
	}
}

// HasUnresolvedOrInvalid returns true if any in-scope validator lacks valid key material.
func (p *RecoveryPlan) HasUnresolvedOrInvalid() bool {
	return p.TotalUnresolved > 0 || p.TotalInvalid > 0
}

// CanCommitByDefault returns true if all in-scope validators were resolved and valid.
func (p *RecoveryPlan) CanCommitByDefault() bool {
	return !p.HasUnresolvedOrInvalid()
}

// GetEntry returns a pointer to the entry for the given pubkey, or nil if not found.
func (p *RecoveryPlan) GetEntry(pubkey types.ValidatorPubkey) *ValidatorPlanEntry {
	for i := range p.Entries {
		if p.Entries[i].Pubkey == pubkey {
			return &p.Entries[i]
		}
	}
	return nil
}

// SummaryString returns a human-readable summary of the plan.
func (p *RecoveryPlan) SummaryString() string {
	return fmt.Sprintf(
		"In-scope: %d (Resolved: %d, Unresolved: %d, Invalid: %d), Excluded: %d",
		p.TotalInScope, p.TotalResolved, p.TotalUnresolved, p.TotalInvalid, p.TotalExcluded,
	)
}

// KeyDeriver provides derived validator keys from the wallet mnemonic.
type KeyDeriver interface {
	GetValidatorKeys(startIndex uint, length uint) ([]wallet.ValidatorKey, error)
}

// InstalledKeyChecker checks if a validator key is already installed in the node.
type InstalledKeyChecker interface {
	IsKeyInstalled(pubkey types.ValidatorPubkey) (bool, error)
}

// CustomKey represents a custom key file that has been parsed.
type CustomKey struct {
	Pubkey         types.ValidatorPubkey
	PrivateKey     *eth2types.BLSPrivateKey
	DerivationPath string
	Error          error
	FileName       string
}

// CustomKeyProvider provides custom validator keys from disk or mocks.
type CustomKeyProvider interface {
	GetCustomKeys() ([]CustomKey, error)
}

// KeyWriter performs actual persistent writes of validator keys.
type KeyWriter interface {
	SaveValidatorKey(key wallet.ValidatorKey) error
	StoreValidatorKey(key *eth2types.BLSPrivateKey, path string) error
}

// Ensure json serialization excludes sensitive fields
var _ json.Marshaler = (*ValidatorPlanEntry)(nil)

func (e ValidatorPlanEntry) MarshalJSON() ([]byte, error) {
	type Alias ValidatorPlanEntry
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(e),
	})
}
