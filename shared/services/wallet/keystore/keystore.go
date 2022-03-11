package keystore

import (
	"github.com/sethvargo/go-password/password"
	eth2types "github.com/wealdtech/go-eth2-types/v2"
)

// Generates a random password
func GenerateRandomPassword() (string, error) {

	// Generate a random 32-character password
	password, err := password.Generate(32, 6, 6, false, false)
	if err != nil {
		return "", err
	}

	return password, nil
}

// Validator keystore interface
type Keystore interface {
	StoreValidatorKey(key *eth2types.BLSPrivateKey, derivationPath string) error
	GetKeystoreDir() string
}
