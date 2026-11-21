package node

import (
	"bytes"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/bindings/node"
	rptypes "github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/rocketpool/validator"
	hexutil "github.com/rocket-pool/smartnode/shared/hex"
	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/services/beacon"
	"github.com/rocket-pool/smartnode/shared/services/passwords"
	"github.com/rocket-pool/smartnode/shared/services/wallet"
	"github.com/rocket-pool/smartnode/shared/services/wallet/keystore"
	lhkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/lighthouse"
	lokeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/lodestar"
	nmkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/nimbus"
	prkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/prysm"
	tkkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/teku"
	"github.com/rocket-pool/smartnode/shared/types/eth2/generic"
)

const testDepositWalletPassword = "correct-horse-battery-staple"

type depositTestFixture struct {
	rootDir      string
	walletPath   string
	passwordPath string
	addressPath  string
	keychainPath string
	pm           *passwords.PasswordManager
	am           *wallet.AddressManager
	wallet       wallet.Wallet
	keystores    map[string]keystore.Keystore
	eth2Config   beacon.Eth2Config
	withdrawCred common.Hash
}

func newDepositTestFixture(t *testing.T, allFiveClients bool) *depositTestFixture {
	t.Helper()

	rootDir := t.TempDir()
	passwordPath := filepath.Join(rootDir, "password")
	walletPath := filepath.Join(rootDir, "wallet")
	addressPath := filepath.Join(rootDir, "address")
	keychainPath := filepath.Join(rootDir, "validators")

	pm := passwords.NewPasswordManager(passwordPath)
	if err := pm.SetPassword(testDepositWalletPassword); err != nil {
		t.Fatalf("SetPassword failed: %v", err)
	}
	am := wallet.NewAddressManager(addressPath)

	w, err := wallet.NewHdWallet(walletPath, 1, big.NewInt(0), big.NewInt(0), 0, pm, am)
	if err != nil {
		t.Fatalf("NewHdWallet failed: %v", err)
	}

	ksMap := map[string]keystore.Keystore{
		"nimbus": nmkeystore.NewKeystore(keychainPath, pm),
		"prysm":  prkeystore.NewKeystore(keychainPath, pm),
	}
	if allFiveClients {
		ksMap["lighthouse"] = lhkeystore.NewKeystore(keychainPath, pm)
		ksMap["lodestar"] = lokeystore.NewKeystore(keychainPath, pm)
		ksMap["teku"] = tkkeystore.NewKeystore(keychainPath, pm)
	}
	for name, ks := range ksMap {
		w.AddKeystore(name, ks)
	}

	if _, err := w.Initialize(wallet.DefaultNodeKeyPath, 0); err != nil {
		t.Fatalf("wallet Initialize failed: %v", err)
	}
	if err := w.Save(); err != nil {
		t.Fatalf("wallet Save failed: %v", err)
	}

	megapoolAddr := common.HexToAddress("0x2222222222222222222222222222222222222222")
	return &depositTestFixture{
		rootDir:      rootDir,
		walletPath:   walletPath,
		passwordPath: passwordPath,
		addressPath:  addressPath,
		keychainPath: keychainPath,
		pm:           pm,
		am:           am,
		wallet:       w,
		keystores:    ksMap,
		eth2Config: beacon.Eth2Config{
			GenesisForkVersion: []byte{0x00, 0x00, 0x00, 0x00},
		},
		withdrawCred: services.CalculateMegapoolWithdrawalCredentials(megapoolAddr),
	}
}

func (f *depositTestFixture) freshDiskKeystores() map[string]keystore.Keystore {
	diskKS := make(map[string]keystore.Keystore, len(f.keystores))
	for name := range f.keystores {
		switch name {
		case "lighthouse":
			diskKS[name] = lhkeystore.NewKeystore(f.keychainPath, f.pm)
		case "lodestar":
			diskKS[name] = lokeystore.NewKeystore(f.keychainPath, f.pm)
		case "nimbus":
			diskKS[name] = nmkeystore.NewKeystore(f.keychainPath, f.pm)
		case "prysm":
			diskKS[name] = prkeystore.NewKeystore(f.keychainPath, f.pm)
		case "teku":
			diskKS[name] = tkkeystore.NewKeystore(f.keychainPath, f.pm)
		}
	}
	return diskKS
}

func (f *depositTestFixture) reloadWalletFromDisk(t *testing.T) wallet.Wallet {
	t.Helper()

	reloaded, err := wallet.NewHdWallet(f.walletPath, 1, big.NewInt(0), big.NewInt(0), 0, f.pm, f.am)
	if err != nil {
		t.Fatalf("reload NewHdWallet failed: %v", err)
	}
	for name, ks := range f.freshDiskKeystores() {
		reloaded.AddKeystore(name, ks)
	}
	return reloaded
}

func (f *depositTestFixture) defaultDeps() depositBatchDeps {
	fourEth := new(big.Int).Mul(big.NewInt(4), big.NewInt(1e18))
	return depositBatchDeps{
		wallet:             f.wallet,
		eth2Config:         f.eth2Config,
		withdrawalCreds:    f.withdrawCred,
		activeValidators:   0,
		nodeBond:           big.NewInt(0),
		nodeQueuedBond:     big.NewInt(0),
		expressTicketCount: 2,
		getBondRequirement: func(validatorCount *big.Int) (*big.Int, error) {
			return new(big.Int).Mul(validatorCount, fourEth), nil
		},
		getDepositData: validator.GetDepositData,
		getValidatorStatus: func(_ rptypes.ValidatorPubkey) (beacon.ValidatorStatus, error) {
			return beacon.ValidatorStatus{Exists: false}, nil
		},
		validateDeposit: validateDepositInfo,
		submitDeposits: func(_ node.Deposits, _ *bind.TransactOpts) (*ethtypes.Transaction, error) {
			return ethtypes.NewTx(&ethtypes.LegacyTx{
				Nonce:    1,
				GasPrice: big.NewInt(1e9),
				Gas:      500000,
				Value:    big.NewInt(0),
			}), nil
		},
	}
}

func (f *depositTestFixture) assertPubkeyAbsentEverywhere(t *testing.T, pubkey rptypes.ValidatorPubkey) {
	t.Helper()

	prefixed := hexutil.AddPrefix(pubkey.Hex())
	candidatePaths := []string{
		filepath.Join(f.keychainPath, lhkeystore.KeystoreDir, lhkeystore.ValidatorsDir, prefixed),
		filepath.Join(f.keychainPath, lhkeystore.KeystoreDir, lhkeystore.SecretsDir, prefixed),
		filepath.Join(f.keychainPath, lokeystore.KeystoreDir, lokeystore.ValidatorsDir, prefixed),
		filepath.Join(f.keychainPath, lokeystore.KeystoreDir, lokeystore.SecretsDir, prefixed),
		filepath.Join(f.keychainPath, nmkeystore.KeystoreDir, nmkeystore.ValidatorsDir, prefixed),
		filepath.Join(f.keychainPath, nmkeystore.KeystoreDir, nmkeystore.SecretsDir, prefixed),
		filepath.Join(f.keychainPath, tkkeystore.KeystoreDir, tkkeystore.ValidatorsDir, prefixed+".json"),
		filepath.Join(f.keychainPath, tkkeystore.KeystoreDir, tkkeystore.SecretsDir, prefixed+".txt"),
	}
	for _, p := range candidatePaths {
		if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("expected %s not to exist for cleaned-up pubkey %s, stat err=%v", p, pubkey.Hex(), err)
		}
	}

	for name, ks := range f.freshDiskKeystores() {
		loaded, err := ks.LoadValidatorKey(pubkey)
		if err != nil {
			t.Fatalf("%s LoadValidatorKey(%s) failed: %v", name, pubkey.Hex(), err)
		}
		if loaded != nil {
			t.Fatalf("%s retained key for cleaned-up pubkey %s", name, pubkey.Hex())
		}
	}
}

func (f *depositTestFixture) assertPubkeyPresentEverywhere(t *testing.T, pubkey rptypes.ValidatorPubkey) {
	t.Helper()

	for name, ks := range f.freshDiskKeystores() {
		loaded, err := ks.LoadValidatorKey(pubkey)
		if err != nil {
			t.Fatalf("%s LoadValidatorKey(%s) failed: %v", name, pubkey.Hex(), err)
		}
		if loaded == nil {
			t.Fatalf("%s missing expected pubkey %s", name, pubkey.Hex())
		}
		gotPub := rptypes.BytesToValidatorPubkey(loaded.PublicKey().Marshal())
		if gotPub != pubkey {
			t.Fatalf("%s returned pubkey %s, want %s", name, gotPub.Hex(), pubkey.Hex())
		}
	}
}

func TestExecuteDepositBatchSuccess(t *testing.T) {
	f := newDepositTestFixture(t, true)
	deps := f.defaultDeps()

	pubkeys, tx, err := executeDepositBatch(deps, 2, 1, true, &bind.TransactOpts{})
	if err != nil {
		t.Fatalf("executeDepositBatch failed: %v", err)
	}
	if tx == nil {
		t.Fatalf("expected non-nil transaction")
	}
	if len(pubkeys) != 2 {
		t.Fatalf("expected 2 validator pubkeys, got %d", len(pubkeys))
	}

	for _, pk := range pubkeys {
		f.assertPubkeyPresentEverywhere(t, pk)
	}

	reloaded := f.reloadWalletFromDisk(t)
	keyCount, err := reloaded.GetValidatorKeyCount()
	if err != nil {
		t.Fatalf("GetValidatorKeyCount failed: %v", err)
	}
	if keyCount != 2 {
		t.Fatalf("expected persisted next_account=2, got %d", keyCount)
	}
}

func TestExecuteDepositBatchCleansUpOnSubmitFailureAndPreservesNextAccount(t *testing.T) {
	f := newDepositTestFixture(t, true)

	// Seed 2 pre-existing active validator keys (indices 0 and 1).
	initialDeps := f.defaultDeps()
	existingPubkeys, _, err := executeDepositBatch(initialDeps, 2, 0, true, &bind.TransactOpts{})
	if err != nil {
		t.Fatalf("initial deposit batch failed: %v", err)
	}

	// Preview indices 2, 3, 4 that will be derived during the failing 3-validator batch.
	upcomingKeys, err := f.wallet.GetValidatorKeys(2, 3)
	if err != nil {
		t.Fatalf("GetValidatorKeys(2, 3) failed: %v", err)
	}

	// Simulate DepositMulti failing (e.g. FeeTooLowToCompete #386, RPC timeout, or contract revert).
	rpcErr := errors.New("replacement transaction underpriced / FeeTooLowToCompete")
	failingDeps := f.defaultDeps()
	failingDeps.activeValidators = 2
	failingDeps.submitDeposits = func(_ node.Deposits, _ *bind.TransactOpts) (*ethtypes.Transaction, error) {
		return nil, rpcErr
	}

	_, _, err = executeDepositBatch(failingDeps, 3, 2, true, &bind.TransactOpts{})
	if !errors.Is(err, rpcErr) {
		t.Fatalf("expected submit error %v, got %v", rpcErr, err)
	}

	// 1. Pre-existing active validator keys must remain untouched across all 5 consensus keystores.
	for _, pk := range existingPubkeys {
		f.assertPubkeyPresentEverywhere(t, pk)
	}

	// 2. All 3 derived keys from the failed batch must be deleted from every keystore on disk.
	burnedSet := map[rptypes.ValidatorPubkey]bool{}
	for _, vk := range upcomingKeys {
		burnedSet[vk.PublicKey] = true
		f.assertPubkeyAbsentEverywhere(t, vk.PublicKey)
	}

	// 3. Reloading a fresh wallet from disk (as services.GetWallet does per CLI call) must preserve next_account=5 (#379).
	reloaded := f.reloadWalletFromDisk(t)
	persistedCount, err := reloaded.GetValidatorKeyCount()
	if err != nil {
		t.Fatalf("reloaded GetValidatorKeyCount failed: %v", err)
	}
	if persistedCount != 5 {
		t.Fatalf("expected persisted next_account=5 (not decremented), got %d", persistedCount)
	}

	// 4. Subsequent retry uses fresh wallet from disk and derives index 5 (never reusing burned indices 2..4, preventing #427).
	retryDeps := f.defaultDeps()
	retryDeps.wallet = reloaded
	retryDeps.activeValidators = 2
	retriedPubkeys, _, err := executeDepositBatch(retryDeps, 1, 0, true, &bind.TransactOpts{})
	if err != nil {
		t.Fatalf("retry executeDepositBatch failed: %v", err)
	}
	if burnedSet[retriedPubkeys[0]] {
		t.Fatalf("retry reused burned pubkey %s instead of advancing past failed indices", retriedPubkeys[0].Hex())
	}
	f.assertPubkeyPresentEverywhere(t, retriedPubkeys[0])

	// 5. Wallet recovery across the burned index gap (2..4) must cleanly recover index 5 and restore keystores.
	keyAt5, err := reloaded.GetValidatorKeyAt(5)
	if err != nil {
		t.Fatalf("GetValidatorKeyAt(5) failed: %v", err)
	}
	if !bytes.Equal(keyAt5.PublicKey().Marshal(), retriedPubkeys[0].Bytes()) {
		t.Fatalf("expected retried validator to be derived at index 5")
	}
	if err := reloaded.DeleteValidatorStores(); err != nil {
		t.Fatalf("DeleteValidatorStores failed: %v", err)
	}
	// Reset in-memory Prysm keystore after DeleteValidatorStores before running full RecoverValidatorKey.
	reloaded.AddKeystore("prysm", prkeystore.NewKeystore(f.keychainPath, f.pm))
	recoveredIdx, err := reloaded.RecoverValidatorKey(retriedPubkeys[0], 0)
	if err != nil {
		t.Fatalf("RecoverValidatorKey across burned gap failed: %v", err)
	}
	if recoveredIdx != 5 {
		t.Fatalf("expected recovered index 5 across burned gap, got %d", recoveredIdx)
	}
	f.assertPubkeyPresentEverywhere(t, retriedPubkeys[0])
}

func TestExecuteDepositBatchMidLoopFailures(t *testing.T) {
	sentinelErr := errors.New("injected mid-batch failure")

	cases := []struct {
		name                 string
		configure            func(deps *depositBatchDeps)
		wantErrSubstr        string
		expectedSavedAccount uint
	}{
		{
			name: "bond_requirement_fails_on_second_validator",
			configure: func(deps *depositBatchDeps) {
				calls := 0
				orig := deps.getBondRequirement
				deps.getBondRequirement = func(validatorCount *big.Int) (*big.Int, error) {
					calls++
					if calls == 2 {
						return nil, sentinelErr
					}
					return orig(validatorCount)
				}
			},
			wantErrSubstr:        sentinelErr.Error(),
			expectedSavedAccount: 1, // validator 1 created; validator 2 failed before CreateValidatorKey
		},
		{
			name: "deposit_data_fails_on_second_validator",
			configure: func(deps *depositBatchDeps) {
				calls := 0
				orig := deps.getDepositData
				deps.getDepositData = func(validatorKey *eth2types.BLSPrivateKey, withdrawalCredentials common.Hash, eth2Config beacon.Eth2Config, depositAmount uint64) (generic.DepositData, common.Hash, error) {
					calls++
					if calls == 2 {
						return generic.DepositData{}, common.Hash{}, sentinelErr
					}
					return orig(validatorKey, withdrawalCredentials, eth2Config, depositAmount)
				}
			},
			wantErrSubstr:        sentinelErr.Error(),
			expectedSavedAccount: 2,
		},
		{
			name: "validator_status_rpc_error_on_second_validator",
			configure: func(deps *depositBatchDeps) {
				calls := 0
				deps.getValidatorStatus = func(_ rptypes.ValidatorPubkey) (beacon.ValidatorStatus, error) {
					calls++
					if calls == 2 {
						return beacon.ValidatorStatus{}, sentinelErr
					}
					return beacon.ValidatorStatus{Exists: false}, nil
				}
			},
			wantErrSubstr:        "Error checking for existing validator status for deposit 2/3",
			expectedSavedAccount: 2,
		},
		{
			name: "duplicate_pubkey_exists_on_beacon_chain",
			configure: func(deps *depositBatchDeps) {
				calls := 0
				deps.getValidatorStatus = func(_ rptypes.ValidatorPubkey) (beacon.ValidatorStatus, error) {
					calls++
					return beacon.ValidatorStatus{Exists: calls == 2}, nil
				}
			},
			wantErrSubstr:        "validator pubkey is already in use on the Beacon chain",
			expectedSavedAccount: 2,
		},
		{
			name: "deposit_signature_validation_fails_on_second_validator",
			configure: func(deps *depositBatchDeps) {
				calls := 0
				orig := deps.validateDeposit
				deps.validateDeposit = func(eth2Config beacon.Eth2Config, depositAmount uint64, pubkey rptypes.ValidatorPubkey, withdrawalCredentials common.Hash, signature rptypes.ValidatorSignature) error {
					calls++
					if calls == 2 {
						return sentinelErr
					}
					return orig(eth2Config, depositAmount, pubkey, withdrawalCredentials, signature)
				}
			},
			wantErrSubstr:        "failed the validation safety check",
			expectedSavedAccount: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newDepositTestFixture(t, false)
			preview, err := f.wallet.GetValidatorKeys(0, 3)
			if err != nil {
				t.Fatalf("GetValidatorKeys failed: %v", err)
			}

			deps := f.defaultDeps()
			tc.configure(&deps)

			_, _, err = executeDepositBatch(deps, 3, 1, true, &bind.TransactOpts{})
			if err == nil || !strings.Contains(err.Error(), tc.wantErrSubstr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErrSubstr, err)
			}

			for _, vk := range preview {
				f.assertPubkeyAbsentEverywhere(t, vk.PublicKey)
			}

			reloaded := f.reloadWalletFromDisk(t)
			gotCount, err := reloaded.GetValidatorKeyCount()
			if err != nil {
				t.Fatalf("GetValidatorKeyCount failed: %v", err)
			}
			if gotCount != tc.expectedSavedAccount {
				t.Fatalf("expected persisted next_account=%d, got %d", tc.expectedSavedAccount, gotCount)
			}
		})
	}
}

func TestExecuteDepositBatchNoSubmitDoesNotMutateWalletFile(t *testing.T) {
	f := newDepositTestFixture(t, false)
	preview, err := f.wallet.GetValidatorKeys(0, 2)
	if err != nil {
		t.Fatalf("GetValidatorKeys failed: %v", err)
	}

	deps := f.defaultDeps()
	simErr := errors.New("offline tx build failed")
	deps.submitDeposits = func(_ node.Deposits, opts *bind.TransactOpts) (*ethtypes.Transaction, error) {
		if !opts.NoSend {
			t.Fatalf("expected opts.NoSend=true when submit=false")
		}
		return nil, simErr
	}

	_, _, err = executeDepositBatch(deps, 2, 0, false, &bind.TransactOpts{})
	if !errors.Is(err, simErr) {
		t.Fatalf("expected %v, got %v", simErr, err)
	}

	for _, vk := range preview {
		f.assertPubkeyAbsentEverywhere(t, vk.PublicKey)
	}

	// Because submit=false never broadcasts a transaction, the persisted wallet file must remain at next_account=0.
	reloaded := f.reloadWalletFromDisk(t)
	gotCount, err := reloaded.GetValidatorKeyCount()
	if err != nil {
		t.Fatalf("GetValidatorKeyCount failed: %v", err)
	}
	if gotCount != 0 {
		t.Fatalf("expected persisted next_account=0 when submit=false fails, got %d", gotCount)
	}
}

func TestExecuteDepositBatchKeepsKeysAfterSuccessfulBroadcastIfWalletSaveFails(t *testing.T) {
	f := newDepositTestFixture(t, false)
	preview, err := f.wallet.GetValidatorKeys(0, 2)
	if err != nil {
		t.Fatalf("GetValidatorKeys failed: %v", err)
	}

	deps := f.defaultDeps()
	deps.submitDeposits = func(deposits node.Deposits, opts *bind.TransactOpts) (*ethtypes.Transaction, error) {
		// Replace wallet file with a directory right as DepositMulti succeeds so the post-broadcast w.Save() fails.
		if err := os.Remove(f.walletPath); err != nil {
			t.Fatalf("Remove walletPath failed: %v", err)
		}
		if err := os.Mkdir(f.walletPath, 0700); err != nil {
			t.Fatalf("Mkdir walletPath failed: %v", err)
		}
		return f.defaultDeps().submitDeposits(deposits, opts)
	}

	_, _, err = executeDepositBatch(deps, 2, 0, true, &bind.TransactOpts{})
	if err == nil {
		t.Fatalf("expected wallet.Save error after walletPath became a directory")
	}

	// CRITICAL SAFETY INVARIANT: once DepositMulti succeeds on-chain, validator keys MUST NOT be deleted!
	for _, vk := range preview {
		f.assertPubkeyPresentEverywhere(t, vk.PublicKey)
	}
}

func TestExecuteDepositBatchCleansUpOnPanic(t *testing.T) {
	f := newDepositTestFixture(t, false)
	preview, err := f.wallet.GetValidatorKeys(0, 2)
	if err != nil {
		t.Fatalf("GetValidatorKeys failed: %v", err)
	}

	deps := f.defaultDeps()
	deps.submitDeposits = func(_ node.Deposits, _ *bind.TransactOpts) (*ethtypes.Transaction, error) {
		panic("unexpected nil pointer in transactor")
	}

	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		_, _, _ = executeDepositBatch(deps, 2, 0, true, &bind.TransactOpts{})
	}()

	if !didPanic {
		t.Fatalf("expected panic to propagate out of executeDepositBatch")
	}

	for _, vk := range preview {
		f.assertPubkeyAbsentEverywhere(t, vk.PublicKey)
	}

	reloaded := f.reloadWalletFromDisk(t)
	gotCount, err := reloaded.GetValidatorKeyCount()
	if err != nil {
		t.Fatalf("GetValidatorKeyCount failed: %v", err)
	}
	if gotCount != 2 {
		t.Fatalf("expected persisted next_account=2 after panic cleanup, got %d", gotCount)
	}
}

func TestExecuteDepositBatchZeroCount(t *testing.T) {
	f := newDepositTestFixture(t, false)
	deps := f.defaultDeps()

	_, _, err := executeDepositBatch(deps, 0, 0, true, &bind.TransactOpts{})
	if err == nil || !strings.Contains(err.Error(), "count must be greater than 0") {
		t.Fatalf("expected zero-count error, got %v", err)
	}
}
