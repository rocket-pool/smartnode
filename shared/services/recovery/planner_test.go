package recovery

import (
	"errors"
	"fmt"
	"testing"

	"github.com/tyler-smith/go-bip39"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	eth2util "github.com/wealdtech/go-eth2-util"

	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
)

const testMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

func init() {
	_ = eth2types.InitBLS()
}

// createTestBLSKey derives a deterministic BLS key for tests using the test mnemonic and index.
func createTestBLSKey(t *testing.T, index uint) (*eth2types.BLSPrivateKey, types.ValidatorPubkey, string) {
	t.Helper()
	seed := bip39.NewSeed(testMnemonic, "")
	derivationPath := fmt.Sprintf("m/12381/3600/%d/0/0", index)
	privKey, err := eth2util.PrivateKeyFromSeedAndPath(seed, derivationPath)
	if err != nil {
		t.Fatalf("failed to derive test BLS key at index %d: %v", index, err)
	}
	pubkeyBytes := privKey.PublicKey().Marshal()
	return privKey, types.BytesToValidatorPubkey(pubkeyBytes), derivationPath
}

// assertZeroWrites ensures no keystore modifications were attempted.
func assertZeroWrites(t *testing.T, writerSpy *MockKeyWriter) {
	t.Helper()
	if writerSpy.TotalWrites() != 0 {
		t.Fatalf("FATAL SAFETY VIOLATION: expected 0 writes on failure, but writer recorded %d writes!", writerSpy.TotalWrites())
	}
	if len(writerSpy.SaveCalls()) != 0 {
		t.Errorf("expected 0 SaveValidatorKey calls, got %d", len(writerSpy.SaveCalls()))
	}
	if len(writerSpy.StoreCalls()) != 0 {
		t.Errorf("expected 0 StoreValidatorKey calls, got %d", len(writerSpy.StoreCalls()))
	}
}

// assertOutcome checks the discovery and commit outcomes for a plan entry.
func assertOutcome(t *testing.T, entry *ValidatorPlanEntry, expectedDiscovery DiscoveryOutcome, expectedCommit CommitOutcome) {
	t.Helper()
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}
	if entry.DiscoveryOutcome != expectedDiscovery {
		t.Errorf("expected DiscoveryOutcome %s for %s, got %s", expectedDiscovery, entry.Pubkey.Hex(), entry.DiscoveryOutcome)
	}
	if entry.CommitOutcome != expectedCommit {
		t.Errorf("expected CommitOutcome %s for %s, got %s", expectedCommit, entry.Pubkey.Hex(), entry.CommitOutcome)
	}
}

// TestMnemonicOnlyRecoverySuccess verifies that a purely mnemonic-derived validator set
// plans cleanly and commits all keys.
func TestMnemonicOnlyRecoverySuccess(t *testing.T) {
	priv0, pk0, path0 := createTestBLSKey(t, 0)
	priv1, pk1, path1 := createTestBLSKey(t, 1)
	priv2, pk2, path2 := createTestBLSKey(t, 2)

	deriver := NewMockKeyDeriver()
	deriver.AddKey(0, wallet.ValidatorKey{PublicKey: pk0, PrivateKey: priv0, DerivationPath: path0, WalletIndex: 0})
	deriver.AddKey(1, wallet.ValidatorKey{PublicKey: pk1, PrivateKey: priv1, DerivationPath: path1, WalletIndex: 1})
	deriver.AddKey(2, wallet.ValidatorKey{PublicKey: pk2, PrivateKey: priv2, DerivationPath: path2, WalletIndex: 2})

	pubkeys := []types.ValidatorPubkey{pk0, pk1, pk2}

	// Step 1: Plan recovery
	plan, err := PlanValidatorRecovery(pubkeys, nil, deriver, nil, nil, DefaultPlannerOptions())
	if err != nil {
		t.Fatalf("unexpected planning error: %v", err)
	}

	if plan.TotalInScope != 3 {
		t.Errorf("expected 3 in-scope keys, got %d", plan.TotalInScope)
	}
	if plan.TotalResolved != 3 {
		t.Errorf("expected 3 resolved keys, got %d", plan.TotalResolved)
	}
	if !plan.CanCommitByDefault() {
		t.Error("plan should allow default commit")
	}

	// Step 2: Commit recovery
	writer := NewMockKeyWriter()
	if err := CommitPlan(plan, writer, false); err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	if writer.TotalWrites() != 3 {
		t.Errorf("expected 3 writes, got %d", writer.TotalWrites())
	}
	for _, entry := range plan.Entries {
		assertOutcome(t, &entry, OutcomeMnemonicValid, CommitWritten)
	}
}

// TestMixedRecoveryWithOneUnresolvedKey_DefaultFlowZeroWrites is the core MVP architectural proof:
// proves that encountering ONE unresolved key halts default recovery and makes ZERO disk writes.
func TestMixedRecoveryWithOneUnresolvedKey_DefaultFlowZeroWrites(t *testing.T) {
	priv0, pk0, path0 := createTestBLSKey(t, 0)
	priv1, pk1, path1 := createTestBLSKey(t, 1)
	privCustom, pkCustom, _ := createTestBLSKey(t, 50)
	_, pkUnresolved, _ := createTestBLSKey(t, 999) // Will not be provided

	deriver := NewMockKeyDeriver()
	deriver.AddKey(0, wallet.ValidatorKey{PublicKey: pk0, PrivateKey: priv0, DerivationPath: path0, WalletIndex: 0})
	deriver.AddKey(1, wallet.ValidatorKey{PublicKey: pk1, PrivateKey: priv1, DerivationPath: path1, WalletIndex: 1})

	customProvider := NewMockCustomKeyProvider([]CustomKey{
		{
			Pubkey:         pkCustom,
			PrivateKey:     privCustom,
			DerivationPath: "custom/keystore-1",
			FileName:       "custom-validator-1.json",
		},
	}, nil)

	pubkeys := []types.ValidatorPubkey{pk0, pk1, pkCustom, pkUnresolved}

	// Step 1: Run read-only plan
	plan, err := PlanValidatorRecovery(pubkeys, nil, deriver, nil, customProvider, DefaultPlannerOptions())
	if err != nil {
		t.Fatalf("unexpected plan error: %v", err)
	}

	if plan.TotalInScope != 4 {
		t.Fatalf("expected 4 in-scope keys, got %d", plan.TotalInScope)
	}
	if plan.TotalResolved != 3 {
		t.Errorf("expected 3 resolved keys, got %d", plan.TotalResolved)
	}
	if plan.TotalUnresolved != 1 {
		t.Errorf("expected 1 unresolved key, got %d", plan.TotalUnresolved)
	}
	if plan.CanCommitByDefault() {
		t.Error("plan with unresolved keys must NOT allow default commit")
	}

	unresolvedEntry := plan.GetEntry(pkUnresolved)
	if unresolvedEntry == nil || unresolvedEntry.DiscoveryOutcome != OutcomeUnresolved {
		t.Fatalf("expected pkUnresolved to have OutcomeUnresolved")
	}

	// Step 2: Attempt commit under DEFAULT mode (allowPartial = false)
	writerSpy := NewMockKeyWriter()
	commitErr := CommitPlan(plan, writerSpy, false)

	if commitErr == nil {
		t.Fatal("expected commit to fail under default mode due to unresolved key, but got nil")
	}
	if !errors.Is(commitErr, ErrUnresolvedKeysRefusingCommit) {
		t.Errorf("expected ErrUnresolvedKeysRefusingCommit, got: %v", commitErr)
	}

	// CRITICAL ARCHITECTURAL GUARANTEE: Zero writes must have been made!
	assertZeroWrites(t, writerSpy)
}

// TestMixedRecoveryWithOneUnresolvedKey_PartialRecoveryAllowsValidKeys verifies that
// operators opting in with allowPartial = true can commit resolved keys while tracking the skipped key.
func TestMixedRecoveryWithOneUnresolvedKey_PartialRecoveryAllowsValidKeys(t *testing.T) {
	priv0, pk0, path0 := createTestBLSKey(t, 0)
	priv1, pk1, path1 := createTestBLSKey(t, 1)
	privCustom, pkCustom, _ := createTestBLSKey(t, 50)
	_, pkUnresolved, _ := createTestBLSKey(t, 999)

	deriver := NewMockKeyDeriver()
	deriver.AddKey(0, wallet.ValidatorKey{PublicKey: pk0, PrivateKey: priv0, DerivationPath: path0, WalletIndex: 0})
	deriver.AddKey(1, wallet.ValidatorKey{PublicKey: pk1, PrivateKey: priv1, DerivationPath: path1, WalletIndex: 1})

	customProvider := NewMockCustomKeyProvider([]CustomKey{
		{
			Pubkey:         pkCustom,
			PrivateKey:     privCustom,
			DerivationPath: "custom/keystore-1",
			FileName:       "custom-validator-1.json",
		},
	}, nil)

	pubkeys := []types.ValidatorPubkey{pk0, pk1, pkCustom, pkUnresolved}

	plan, err := PlanValidatorRecovery(pubkeys, nil, deriver, nil, customProvider, DefaultPlannerOptions())
	if err != nil {
		t.Fatalf("unexpected plan error: %v", err)
	}

	writerSpy := NewMockKeyWriter()
	// Commit with allowPartial = true
	if err := CommitPlan(plan, writerSpy, true); err != nil {
		t.Fatalf("unexpected commit error with partial recovery allowed: %v", err)
	}

	// Should have written the 3 resolved keys (2 mnemonic + 1 custom)
	if writerSpy.TotalWrites() != 3 {
		t.Errorf("expected 3 writes, got %d", writerSpy.TotalWrites())
	}
	if len(writerSpy.SaveCalls()) != 2 {
		t.Errorf("expected 2 SaveValidatorKey calls, got %d", len(writerSpy.SaveCalls()))
	}
	if len(writerSpy.StoreCalls()) != 1 {
		t.Errorf("expected 1 StoreValidatorKey calls, got %d", len(writerSpy.StoreCalls()))
	}

	unresolvedEntry := plan.GetEntry(pkUnresolved)
	if unresolvedEntry.CommitOutcome != CommitSkippedPartial {
		t.Errorf("expected unresolved entry to have CommitSkippedPartial, got %s", unresolvedEntry.CommitOutcome)
	}
}

// TestInactiveValidatorFiltering_FixIssue407 verifies that inactive and exited validators
// are excluded cleanly during planning and do not block recovery or generate missing-key errors.
func TestInactiveValidatorFiltering_FixIssue407(t *testing.T) {
	privActive, pkActive, pathActive := createTestBLSKey(t, 0)
	privPending, pkPending, pathPending := createTestBLSKey(t, 1)
	_, pkExited, _ := createTestBLSKey(t, 2)
	_, pkSlashed, _ := createTestBLSKey(t, 3)

	deriver := NewMockKeyDeriver()
	deriver.AddKey(0, wallet.ValidatorKey{PublicKey: pkActive, PrivateKey: privActive, DerivationPath: pathActive, WalletIndex: 0})
	deriver.AddKey(1, wallet.ValidatorKey{PublicKey: pkPending, PrivateKey: privPending, DerivationPath: pathPending, WalletIndex: 1})

	statuses := map[types.ValidatorPubkey]beacon.ValidatorStatus{
		pkActive:  {Status: beacon.ValidatorState_ActiveOngoing},
		pkPending: {Status: beacon.ValidatorState_PendingQueued},
		pkExited:  {Status: beacon.ValidatorState_ExitedUnslashed},
		pkSlashed: {Status: beacon.ValidatorState_ExitedSlashed},
	}

	pubkeys := []types.ValidatorPubkey{pkActive, pkPending, pkExited, pkSlashed}

	plan, err := PlanValidatorRecovery(pubkeys, statuses, deriver, nil, nil, DefaultPlannerOptions())
	if err != nil {
		t.Fatalf("unexpected plan error: %v", err)
	}

	if plan.TotalExcluded != 2 {
		t.Errorf("expected 2 excluded inactive validators, got %d", plan.TotalExcluded)
	}
	if plan.TotalInScope != 2 {
		t.Errorf("expected 2 in-scope validators, got %d", plan.TotalInScope)
	}
	if plan.TotalResolved != 2 {
		t.Errorf("expected 2 resolved validators, got %d", plan.TotalResolved)
	}
	if plan.TotalUnresolved != 0 {
		t.Errorf("expected 0 unresolved validators, got %d", plan.TotalUnresolved)
	}
	if !plan.CanCommitByDefault() {
		t.Error("plan with excluded inactive validators should be committable by default")
	}

	exitedEntry := plan.GetEntry(pkExited)
	assertOutcome(t, exitedEntry, OutcomeExcludedInactive, CommitSkippedExcluded)

	writer := NewMockKeyWriter()
	if err := CommitPlan(plan, writer, false); err != nil {
		t.Fatalf("unexpected commit error: %v", err)
	}

	if writer.TotalWrites() != 2 {
		t.Errorf("expected exactly 2 writes for the 2 active validators, got %d", writer.TotalWrites())
	}
}

// TestCustomKeyInvalidMaterial_FailsFast verifies that corrupt keystores or password mismatches
// are captured as OutcomeInvalidMaterial and prevent default writes.
func TestCustomKeyInvalidMaterial_FailsFast(t *testing.T) {
	_, pkBad, _ := createTestBLSKey(t, 0)

	customProvider := NewMockCustomKeyProvider([]CustomKey{
		{
			Pubkey:   pkBad,
			Error:    errors.New("could not decrypt keystore: invalid password"),
			FileName: "bad-keystore.json",
		},
	}, nil)

	pubkeys := []types.ValidatorPubkey{pkBad}

	plan, err := PlanValidatorRecovery(pubkeys, nil, nil, nil, customProvider, DefaultPlannerOptions())
	if err != nil {
		t.Fatalf("unexpected planning error: %v", err)
	}

	if plan.TotalInvalid != 1 {
		t.Errorf("expected 1 invalid key, got %d", plan.TotalInvalid)
	}
	if plan.CanCommitByDefault() {
		t.Error("plan with invalid custom key material must NOT allow default commit")
	}

	writer := NewMockKeyWriter()
	err = CommitPlan(plan, writer, false)
	if err == nil {
		t.Fatal("expected commit to fail on invalid custom key, got nil")
	}
	assertZeroWrites(t, writer)
}

// TestAlreadyInstalledKey_SkippedDuringCommit verifies that keys already present in the validator client
// are diagnosed as installed and skipped during commit without error.
func TestAlreadyInstalledKey_SkippedDuringCommit(t *testing.T) {
	_, pkInstalled, _ := createTestBLSKey(t, 0)
	privNew, pkNew, pathNew := createTestBLSKey(t, 1)

	deriver := NewMockKeyDeriver()
	deriver.AddKey(1, wallet.ValidatorKey{PublicKey: pkNew, PrivateKey: privNew, DerivationPath: pathNew, WalletIndex: 1})

	installedChecker := NewMockInstalledKeyChecker(pkInstalled)

	pubkeys := []types.ValidatorPubkey{pkInstalled, pkNew}

	plan, err := PlanValidatorRecovery(pubkeys, nil, deriver, installedChecker, nil, DefaultPlannerOptions())
	if err != nil {
		t.Fatalf("unexpected plan error: %v", err)
	}

	installedEntry := plan.GetEntry(pkInstalled)
	assertOutcome(t, installedEntry, OutcomeInstalledValid, CommitSkippedAlreadyInstalled)

	writer := NewMockKeyWriter()
	if err := CommitPlan(plan, writer, false); err != nil {
		t.Fatalf("unexpected commit error: %v", err)
	}

	// Only the new key pkNew should have been written!
	if writer.TotalWrites() != 1 {
		t.Errorf("expected 1 write, got %d", writer.TotalWrites())
	}
	saveCalls := writer.SaveCalls()
	if len(saveCalls) != 1 || saveCalls[0].PublicKey != pkNew {
		t.Errorf("expected SaveValidatorKey call for pkNew, got %+v", saveCalls)
	}
}
