package wallet

import (
	"bytes"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	eth2types "github.com/wealdtech/go-eth2-types/v2"

	"github.com/rocket-pool/smartnode/bindings/types"
	"github.com/rocket-pool/smartnode/rocketpool/validator"
	hexutil "github.com/rocket-pool/smartnode/shared/hex"
	"github.com/rocket-pool/smartnode/shared/services/passwords"
	"github.com/rocket-pool/smartnode/shared/services/wallet/keystore"
	lhkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/lighthouse"
	lokeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/lodestar"
	nmkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/nimbus"
	prkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/prysm"
	tkkeystore "github.com/rocket-pool/smartnode/shared/services/wallet/keystore/teku"
)

const testWalletPassword = "correct-horse-battery-staple"

type testWalletEnv struct {
	rootDir      string
	walletPath   string
	passwordPath string
	addressPath  string
	keychainPath string
	pm           *passwords.PasswordManager
	am           *AddressManager
	wallet       Wallet
	keystores    map[string]keystore.Keystore
}

func newInitializedTestWallet(t *testing.T, registerDefaultKeystores bool) *testWalletEnv {
	t.Helper()

	rootDir := t.TempDir()
	passwordPath := filepath.Join(rootDir, "password")
	walletPath := filepath.Join(rootDir, "wallet")
	addressPath := filepath.Join(rootDir, "address")
	keychainPath := filepath.Join(rootDir, "validators")

	pm := passwords.NewPasswordManager(passwordPath)
	if err := pm.SetPassword(testWalletPassword); err != nil {
		t.Fatalf("failed to set test wallet password: %v", err)
	}
	am := NewAddressManager(addressPath)

	w, err := NewHdWallet(walletPath, 1, big.NewInt(0), big.NewInt(0), 0, pm, am)
	if err != nil {
		t.Fatalf("NewHdWallet failed: %v", err)
	}

	ksMap := map[string]keystore.Keystore{}
	if registerDefaultKeystores {
		ksMap["lighthouse"] = lhkeystore.NewKeystore(keychainPath, pm)
		ksMap["lodestar"] = lokeystore.NewKeystore(keychainPath, pm)
		ksMap["nimbus"] = nmkeystore.NewKeystore(keychainPath, pm)
		ksMap["prysm"] = prkeystore.NewKeystore(keychainPath, pm)
		ksMap["teku"] = tkkeystore.NewKeystore(keychainPath, pm)
		for name, ks := range ksMap {
			w.AddKeystore(name, ks)
		}
	}

	if _, err := w.Initialize(DefaultNodeKeyPath, 0); err != nil {
		t.Fatalf("wallet Initialize failed: %v", err)
	}
	if err := w.Save(); err != nil {
		t.Fatalf("wallet Save failed: %v", err)
	}

	return &testWalletEnv{
		rootDir:      rootDir,
		walletPath:   walletPath,
		passwordPath: passwordPath,
		addressPath:  addressPath,
		keychainPath: keychainPath,
		pm:           pm,
		am:           am,
		wallet:       w,
		keystores:    ksMap,
	}
}

func assertKeyAbsentOnDiskAndKeystores(t *testing.T, env *testWalletEnv, pubkey types.ValidatorPubkey) {
	t.Helper()

	prefixed := hexutil.AddPrefix(pubkey.Hex())
	perKeyArtifacts := []string{
		filepath.Join(env.keychainPath, lhkeystore.KeystoreDir, lhkeystore.ValidatorsDir, prefixed),
		filepath.Join(env.keychainPath, lhkeystore.KeystoreDir, lhkeystore.SecretsDir, prefixed),
		filepath.Join(env.keychainPath, lokeystore.KeystoreDir, lokeystore.ValidatorsDir, prefixed),
		filepath.Join(env.keychainPath, lokeystore.KeystoreDir, lokeystore.SecretsDir, prefixed),
		filepath.Join(env.keychainPath, nmkeystore.KeystoreDir, nmkeystore.ValidatorsDir, prefixed),
		filepath.Join(env.keychainPath, nmkeystore.KeystoreDir, nmkeystore.SecretsDir, prefixed),
		filepath.Join(env.keychainPath, tkkeystore.KeystoreDir, tkkeystore.ValidatorsDir, prefixed+".json"),
		filepath.Join(env.keychainPath, tkkeystore.KeystoreDir, tkkeystore.SecretsDir, prefixed+".txt"),
	}
	for _, p := range perKeyArtifacts {
		if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("expected artifact %s to be deleted for %s, stat err=%v", p, pubkey.Hex(), err)
		}
	}

	for name, ks := range env.keystores {
		loaded, err := ks.LoadValidatorKey(pubkey)
		if err != nil {
			t.Fatalf("%s LoadValidatorKey(%s) returned unexpected error: %v", name, pubkey.Hex(), err)
		}
		if loaded != nil {
			t.Fatalf("%s still returned validator key for deleted pubkey %s", name, pubkey.Hex())
		}
	}

	// Reload a fresh Prysm keystore from disk to verify the persisted encrypted store also omits pubkey.
	reloadedPrysm := prkeystore.NewKeystore(env.keychainPath, env.pm)
	loadedFromDisk, err := reloadedPrysm.LoadValidatorKey(pubkey)
	if err != nil {
		t.Fatalf("reloaded prysm LoadValidatorKey(%s) returned error: %v", pubkey.Hex(), err)
	}
	if loadedFromDisk != nil {
		t.Fatalf("reloaded prysm keystore still contains deleted pubkey %s on disk", pubkey.Hex())
	}
}

func assertKeyPresentInAllKeystores(t *testing.T, env *testWalletEnv, expected *eth2types.BLSPrivateKey) {
	t.Helper()

	pubkey := types.BytesToValidatorPubkey(expected.PublicKey().Marshal())
	for name, ks := range env.keystores {
		loaded, err := ks.LoadValidatorKey(pubkey)
		if err != nil {
			t.Fatalf("%s LoadValidatorKey(%s) failed: %v", name, pubkey.Hex(), err)
		}
		if loaded == nil {
			t.Fatalf("%s missing expected validator key %s", name, pubkey.Hex())
		}
		if !bytes.Equal(loaded.Marshal(), expected.Marshal()) {
			t.Fatalf("%s returned mismatched private key bytes for %s", name, pubkey.Hex())
		}
	}

	reloadedPrysm := prkeystore.NewKeystore(env.keychainPath, env.pm)
	loadedFromDisk, err := reloadedPrysm.LoadValidatorKey(pubkey)
	if err != nil {
		t.Fatalf("reloaded prysm LoadValidatorKey(%s) failed: %v", pubkey.Hex(), err)
	}
	if loadedFromDisk == nil || !bytes.Equal(loadedFromDisk.Marshal(), expected.Marshal()) {
		t.Fatalf("reloaded prysm missing or mismatched key for %s", pubkey.Hex())
	}
}

func TestDeleteValidatorKeyAcrossAllKeystores(t *testing.T) {
	env := newInitializedTestWallet(t, true)

	keys := make([]*eth2types.BLSPrivateKey, 3)
	pubkeys := make([]types.ValidatorPubkey, 3)
	for i := range keys {
		key, err := env.wallet.CreateValidatorKey()
		if err != nil {
			t.Fatalf("CreateValidatorKey(%d) failed: %v", i, err)
		}
		keys[i] = key
		pubkeys[i] = types.BytesToValidatorPubkey(key.PublicKey().Marshal())
	}

	// Delete the middle validator key and verify neighbors remain intact across all 5 keystores.
	if err := env.wallet.DeleteValidatorKey(pubkeys[1]); err != nil {
		t.Fatalf("DeleteValidatorKey(pubkeys[1]) failed: %v", err)
	}

	assertKeyAbsentOnDiskAndKeystores(t, env, pubkeys[1])
	assertKeyPresentInAllKeystores(t, env, keys[0])
	assertKeyPresentInAllKeystores(t, env, keys[2])

	// Deleting the same key a second time must be idempotent.
	if err := env.wallet.DeleteValidatorKey(pubkeys[1]); err != nil {
		t.Fatalf("second DeleteValidatorKey(pubkeys[1]) failed: %v", err)
	}

	// next_account must NOT be decremented when deleting a derived key (#379).
	count, err := env.wallet.GetValidatorKeyCount()
	if err != nil {
		t.Fatalf("GetValidatorKeyCount failed: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected next_account=3 after DeleteValidatorKey, got %d", count)
	}
}

func TestPrysmDeleteLastKeyAndReuse(t *testing.T) {
	if err := validator.InitializeBLS(); err != nil {
		t.Fatalf("InitializeBLS failed: %v", err)
	}

	env := newInitializedTestWallet(t, false)
	prysmKS := prkeystore.NewKeystore(env.keychainPath, env.pm)
	env.wallet.AddKeystore("prysm", prysmKS)

	key0, err := env.wallet.CreateValidatorKey()
	if err != nil {
		t.Fatalf("CreateValidatorKey(0) failed: %v", err)
	}
	pk0 := types.BytesToValidatorPubkey(key0.PublicKey().Marshal())

	if err := env.wallet.DeleteValidatorKey(pk0); err != nil {
		t.Fatalf("DeleteValidatorKey(pk0) failed: %v", err)
	}

	reloaded := prkeystore.NewKeystore(env.keychainPath, env.pm)
	loaded0, err := reloaded.LoadValidatorKey(pk0)
	if err != nil {
		t.Fatalf("reloaded LoadValidatorKey(pk0) failed: %v", err)
	}
	if loaded0 != nil {
		t.Fatalf("expected deleted pk0 to be absent from reloaded Prysm keystore")
	}

	// Creating a subsequent key on the emptied Prysm store must succeed cleanly.
	key1, err := env.wallet.CreateValidatorKey()
	if err != nil {
		t.Fatalf("CreateValidatorKey(1) failed: %v", err)
	}
	pk1 := types.BytesToValidatorPubkey(key1.PublicKey().Marshal())

	reloadedAfterReuse := prkeystore.NewKeystore(env.keychainPath, env.pm)
	loaded1, err := reloadedAfterReuse.LoadValidatorKey(pk1)
	if err != nil {
		t.Fatalf("reloaded LoadValidatorKey(pk1) failed: %v", err)
	}
	if loaded1 == nil || !bytes.Equal(loaded1.Marshal(), key1.Marshal()) {
		t.Fatalf("expected newly stored pk1 to load from Prysm keystore after emptying")
	}
}

func TestDeleteValidatorKeyIdempotentOnEmptyAndMissing(t *testing.T) {
	env := newInitializedTestWallet(t, true)

	nextKey, err := env.wallet.GetNextValidatorKey()
	if err != nil {
		t.Fatalf("GetNextValidatorKey failed: %v", err)
	}
	missingPubkey := types.BytesToValidatorPubkey(nextKey.PublicKey().Marshal())

	if err := env.wallet.DeleteValidatorKey(missingPubkey); err != nil {
		t.Fatalf("DeleteValidatorKey on uninitialized keystores failed: %v", err)
	}

	// Deleting from an untouched wallet must not initialize directories or create Prysm secret files.
	if _, err := os.Stat(env.keychainPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected keychainPath %s not to be created on no-op delete, stat err=%v", env.keychainPath, err)
	}
}

type stubKeystore struct {
	dir      string
	storeErr error
	delErr   error
	stored   map[types.ValidatorPubkey]*eth2types.BLSPrivateKey
}

func (s *stubKeystore) StoreValidatorKey(key *eth2types.BLSPrivateKey, _ string) error {
	if s.storeErr != nil {
		return s.storeErr
	}
	if s.stored == nil {
		s.stored = map[types.ValidatorPubkey]*eth2types.BLSPrivateKey{}
	}
	pk := types.BytesToValidatorPubkey(key.PublicKey().Marshal())
	s.stored[pk] = key
	return nil
}

func (s *stubKeystore) LoadValidatorKey(pubkey types.ValidatorPubkey) (*eth2types.BLSPrivateKey, error) {
	if s.stored == nil {
		return nil, nil
	}
	return s.stored[pubkey], nil
}

func (s *stubKeystore) DeleteValidatorKey(pubkey types.ValidatorPubkey) error {
	if s.delErr != nil {
		return s.delErr
	}
	delete(s.stored, pubkey)
	return nil
}

func (s *stubKeystore) GetKeystoreDir() string {
	return s.dir
}

func TestCreateValidatorKeyRollsBackPartialKeystoreFailure(t *testing.T) {
	env := newInitializedTestWallet(t, false)

	nimbusKS := nmkeystore.NewKeystore(env.keychainPath, env.pm)
	prysmKS := prkeystore.NewKeystore(env.keychainPath, env.pm)
	healthyStub := &stubKeystore{dir: filepath.Join(env.keychainPath, "healthy")}
	failingErr := errors.New("disk full while writing keystore")
	failingStub := &stubKeystore{
		dir:      filepath.Join(env.keychainPath, "failing"),
		storeErr: failingErr,
	}

	env.wallet.AddKeystore("nimbus", nimbusKS)
	env.wallet.AddKeystore("prysm", prysmKS)
	env.wallet.AddKeystore("healthy", healthyStub)
	env.wallet.AddKeystore("failing", failingStub)

	previewKey, err := env.wallet.GetNextValidatorKey()
	if err != nil {
		t.Fatalf("GetNextValidatorKey failed: %v", err)
	}
	previewPubkey := types.BytesToValidatorPubkey(previewKey.PublicKey().Marshal())

	_, err = env.wallet.CreateValidatorKey()
	if !errors.Is(err, failingErr) {
		t.Fatalf("expected CreateValidatorKey error wrapping %v, got %v", failingErr, err)
	}

	// Partial writes to healthy keystores must have been rolled back.
	for name, ks := range map[string]keystore.Keystore{
		"nimbus":  nimbusKS,
		"prysm":   prysmKS,
		"healthy": healthyStub,
	} {
		loaded, loadErr := ks.LoadValidatorKey(previewPubkey)
		if loadErr != nil {
			t.Fatalf("%s LoadValidatorKey failed: %v", name, loadErr)
		}
		if loaded != nil {
			t.Fatalf("%s leaked validator key %s after partial StoreValidatorKey failure", name, previewPubkey.Hex())
		}
	}

	// next_account remains incremented so burned index 0 is not reused.
	count, err := env.wallet.GetValidatorKeyCount()
	if err != nil {
		t.Fatalf("GetValidatorKeyCount failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected next_account=1 after failed CreateValidatorKey, got %d", count)
	}
}

func TestDeleteValidatorKeyContinuesAfterKeystoreError(t *testing.T) {
	env := newInitializedTestWallet(t, false)

	nimbusKS := nmkeystore.NewKeystore(env.keychainPath, env.pm)
	prysmKS := prkeystore.NewKeystore(env.keychainPath, env.pm)
	deleteBoom := errors.New("permission denied removing keystore file")
	brokenKS := &stubKeystore{
		dir:    filepath.Join(env.keychainPath, "aa-broken"),
		delErr: deleteBoom,
	}

	// Name "aa-broken" sorts before "nimbus" and "prysm" to prove DeleteValidatorKey continues past errors.
	env.wallet.AddKeystore("aa-broken", brokenKS)
	env.wallet.AddKeystore("nimbus", nimbusKS)
	env.wallet.AddKeystore("prysm", prysmKS)

	key, err := env.wallet.CreateValidatorKey()
	if err != nil {
		t.Fatalf("CreateValidatorKey failed: %v", err)
	}
	pk := types.BytesToValidatorPubkey(key.PublicKey().Marshal())

	err = env.wallet.DeleteValidatorKey(pk)
	if !errors.Is(err, deleteBoom) {
		t.Fatalf("expected DeleteValidatorKey to join %v, got %v", deleteBoom, err)
	}

	for name, ks := range map[string]keystore.Keystore{
		"nimbus": nimbusKS,
		"prysm":  prysmKS,
	} {
		loaded, loadErr := ks.LoadValidatorKey(pk)
		if loadErr != nil {
			t.Fatalf("%s LoadValidatorKey failed: %v", name, loadErr)
		}
		if loaded != nil {
			t.Fatalf("%s was not cleaned up after earlier keystore error", name)
		}
	}
}

func TestMasqueradeWalletDeleteValidatorKey(t *testing.T) {
	env := newInitializedTestWallet(t, false)
	if err := env.am.SetAndSaveAddress(common.HexToAddress("0x1111111111111111111111111111111111111111"), true); err != nil {
		t.Fatalf("SetAndSaveAddress failed: %v", err)
	}

	mw, err := NewWallet(env.addressPath, env.walletPath, 1, big.NewInt(0), big.NewInt(0), 0, env.pm, env.am)
	if err != nil {
		t.Fatalf("NewWallet (masquerade) failed: %v", err)
	}

	var dummyPubkey types.ValidatorPubkey
	if err := mw.DeleteValidatorKey(dummyPubkey); !errors.Is(err, ErrIsMasquerading) {
		t.Fatalf("expected ErrIsMasquerading from masqueradeWallet.DeleteValidatorKey, got %v", err)
	}
}
