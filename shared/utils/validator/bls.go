package validator

import (
	"fmt"
	"sync"

	"github.com/tyler-smith/go-bip39"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
	eth2util "github.com/wealdtech/go-eth2-util"
)

// Config
const (
	ValidatorKeyPath string = "m/12381/3600/%d/0/0"
)

// BLS signing root with domain
type signingRoot struct {
	ObjectRoot []byte `ssz-size:"32"`
	Domain     []byte `ssz-size:"32"`
}

// Initialize BLS support
var initBLS sync.Once

func InitializeBLS() error {
	var err error
	initBLS.Do(func() {
		err = eth2types.InitBLS()
	})
	return err
}

// Get a private BLS key from the mnemonic, index, and path
func GetPrivateKey(mnemonic string, index uint, path string) (*eth2types.BLSPrivateKey, error) {

	// Get derivation path
	derivationPath := fmt.Sprintf(path, index)

	// Generate seed
	seed := bip39.NewSeed(mnemonic, "")

	// Initialize BLS support
	if err := InitializeBLS(); err != nil {
		return nil, fmt.Errorf("Could not initialize BLS library: %w", err)
	}

	// Get private key
	privateKey, err := eth2util.PrivateKeyFromSeedAndPath(seed, derivationPath)
	if err != nil {
		return nil, fmt.Errorf("Could not get validator %d private key: %w", index, err)
	}

	// Return
	return privateKey, nil

}
